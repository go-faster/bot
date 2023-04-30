package stat

import (
	"context"
	"fmt"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v52/github"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/ent/gitcommit"
)

func NewCommit(
	db *ent.Client,
	cache *redis.Client,
	gh *github.Client,
	ghID int64,
	meterProvider metric.MeterProvider,
	tracerProvider trace.TracerProvider,
) *Commit {
	const prefix = "github.com/go-faster/bot/internal/stat.Commit"
	meter := meterProvider.Meter(prefix)
	return &Commit{
		cache:    cache,
		db:       db,
		ghClient: gh,
		ghID:     ghID,
		meter:    meter,
		tracer:   tracerProvider.Tracer(prefix),
	}
}

type Commit struct {
	db    *ent.Client
	cache *redis.Client

	ghID     int64
	ghClient *github.Client

	tracer trace.Tracer
	meter  metric.Meter
}

func (w *Commit) Update(ctx context.Context) error {
	ctx, span := w.tracer.Start(ctx, "Update")
	defer span.End()

	client, err := w.client(ctx)
	if err != nil {
		return errors.Wrap(err, "client")
	}

	tx, err := w.db.Tx(ctx)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}

	all, err := tx.Repository.Query().WithOrganization().All(ctx)
	if err != nil {
		return errors.Wrap(err, "query repositories")
	}

	for _, repo := range all {
		commits, _, err := client.Repositories.ListCommits(ctx, repo.Edges.Organization.Name, repo.Name, &github.CommitsListOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		})
		if err != nil {
			return errors.Wrap(err, "list commits")
		}
		for _, commit := range commits {
			zctx.From(ctx).Info("Commit",
				zap.String("sha", commit.GetSHA()),
				zap.String("message", commit.GetCommit().GetMessage()),
				zap.String("repo", repo.FullName),
			)
			if err := tx.GitCommit.Create().
				SetID(commit.GetSHA()).
				SetDate(commit.GetCommit().GetAuthor().GetDate().Time).
				SetAuthorLogin(commit.GetAuthor().GetLogin()).
				SetAuthorID(commit.GetAuthor().GetID()).
				SetMessage(commit.GetCommit().GetMessage()).
				OnConflictColumns(gitcommit.FieldID).Ignore().Exec(ctx); err != nil {
				return errors.Wrap(err, "create commit")
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit")
	}

	return nil
}

func (w *Commit) clientWithToken(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func (w *Commit) client(ctx context.Context) (*github.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	installations, _, err := w.ghClient.Apps.ListInstallations(ctx, &github.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "list installations")
	}
	for _, inst := range installations {
		zctx.From(ctx).Info("Installation",
			zap.Int64("id", inst.GetID()),
			zap.String("account", inst.GetAccount().GetLogin()),
		)
	}

	tokenKey := fmt.Sprintf("gh:installation:%d:token", w.ghID)
	key, err := w.cache.Get(ctx, tokenKey).Result()
	if err == nil {
		return w.clientWithToken(ctx, key), nil
	}

	tok, _, err := w.ghClient.Apps.CreateInstallationToken(ctx, w.ghID, &github.InstallationTokenOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "create token")
	}

	expiration := time.Until(tok.GetExpiresAt().Time)
	zctx.From(ctx).Info("Token expires in",
		zap.Duration("d", expiration),
	)
	offset := time.Minute * 10
	if expiration > offset {
		// Just to make sure that we will not get expired token.
		expiration -= offset
	}
	if _, err := w.cache.Set(ctx, tokenKey, tok.GetToken(), expiration).Result(); err != nil {
		return nil, errors.Wrap(err, "set token")
	}

	return w.clientWithToken(ctx, tok.GetToken()), nil
}
