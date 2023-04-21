package gh

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/multierr"

	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/ent/check"
	"github.com/go-faster/bot/internal/ent/lastchannelmessage"
	"github.com/go-faster/bot/internal/ent/prnotification"
)

// Hook is event handler which saves last message ID of dialog to the storage.
type Hook struct {
	next dispatch.MessageHandler
	db   *ent.LastChannelMessageClient
}

// NewHook creates new hook.
func NewHook(next dispatch.MessageHandler, db *ent.LastChannelMessageClient) Hook {
	return Hook{next: next, db: db}
}

// OnMessage implements dispatch.MessageHandler.
func (h Hook) OnMessage(ctx context.Context, e dispatch.MessageEvent) error {
	ch, ok := e.Channel()
	if !ok {
		return h.next.OnMessage(ctx, e)
	}

	return multierr.Append(
		updateLastMsgID(ctx, h.db, ch.ID, e.Message.ID),
		h.next.OnMessage(ctx, e),
	)
}

func updateLastMsgID(ctx context.Context, db *ent.LastChannelMessageClient, channelID int64, msgID int) error {
	// FIXME(tdakkota): tracing
	if err := db.Create().
		SetID(channelID).
		SetMessageID(msgID).
		OnConflict(
			sql.ConflictColumns(lastchannelmessage.FieldID),
			sql.ResolveWithNewValues(),
		).
		UpdateMessageID().
		Exec(ctx); err != nil {
		return errors.Wrap(err, "upsert last message")
	}

	return nil
}

func (h *Webhook) updateLastMsgID(ctx context.Context, channelID int64, msgID int) error {
	return updateLastMsgID(ctx, h.db.LastChannelMessage, channelID, msgID)
}

func (h *Webhook) findPRNotification(
	ctx context.Context,
	repo *github.Repository,
	pr *github.PullRequest,
	channelID int64,
) (existingMsgID, lastMsgID int, _ error) {
	ctx, span := h.tracer.Start(ctx, "FindPRNotification",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	tx, err := h.db.Tx(ctx)
	if err != nil {
		return 0, 0, errors.Wrap(err, "begin tx")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	switch n, err := tx.PRNotification.Query().
		Where(
			prnotification.PullRequestIDEQ(pr.GetNumber()),
			prnotification.RepoIDEQ(repo.GetID()),
		).
		First(ctx); {
	case err == nil:
		existingMsgID = n.MessageID
	case ent.IsNotFound(err):
	default:
		return 0, 0, errors.Wrap(err, "query pr notification")
	}

	switch m, err := tx.LastChannelMessage.Query().
		Where(lastchannelmessage.IDEQ(channelID)).
		First(ctx); {
	case err == nil:
		lastMsgID = m.MessageID
	case ent.IsNotFound(err):
	default:
		return 0, 0, errors.Wrap(err, "query last message")
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, errors.Wrap(err, "commit")
	}

	return existingMsgID, lastMsgID, nil
}

func (h *Webhook) setPRNotification(ctx context.Context, repo *github.Repository, pr *github.PullRequest, msgID int) error {
	ctx, span := h.tracer.Start(ctx, "SetPRNotification",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	u := pr.GetUser()
	if err := h.db.PRNotification.Create().
		SetRepoID(repo.GetID()).
		SetPullRequestID(pr.GetNumber()).
		SetPullRequestTitle(pr.GetTitle()).
		SetPullRequestBody(pr.GetBody()).
		SetPullRequestAuthorLogin(u.GetLogin()).
		SetMessageID(msgID).
		OnConflict(
			sql.ConflictColumns(
				prnotification.FieldRepoID,
				prnotification.FieldPullRequestID,
			),
			sql.ResolveWithNewValues(),
		).
		UpdateNewValues().
		Exec(ctx); err != nil {
		return errors.Wrap(err, "upsert PR notification")
	}

	return nil
}

// fillPRState queries misisng data from database if needed.
func (h *Webhook) fillPRState(ctx context.Context, tx *ent.PRNotificationClient, repo *github.Repository, pr *github.PullRequest) error {
	if pr.GetHTMLURL() == "" {
		u := fmt.Sprintf("https://github.com/%s/%s/pull/%d",
			repo.GetOwner().GetLogin(), repo.GetName(),
			pr.GetNumber(),
		)
		pr.HTMLURL = &u
	}

	author := pr.GetUser()
	needQuery := pr.GetTitle() == "" ||
		author.GetLogin() == ""

	if !needQuery {
		return nil
	}

	cached, err := tx.Query().
		Where(
			prnotification.RepoID(repo.GetID()),
			prnotification.PullRequestID(pr.GetNumber()),
		).First(ctx)
	if err != nil {
		return err
	}

	if pr.GetTitle() == "" {
		pr.Title = &cached.PullRequestTitle
	}

	if author.GetLogin() == "" {
		if pr.User == nil {
			pr.User = new(github.User)
		}
		pr.User.Login = &cached.PullRequestAuthorLogin
	}

	return nil
}

type Check struct {
	ID         int64
	Name       string // e.g. "build"
	Conclusion string // failure, success, neutral, cancelled, skipped, timed_out, action_required
	Status     string // completed
}

func (h *Webhook) queryChecks(ctx context.Context, repo *github.Repository, pr *github.PullRequest) (checks []Check, _ error) {
	ctx, span := h.tracer.Start(ctx, "QueryChecks",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("repository.full_name", repo.GetFullName()),
			attribute.Int64("repository.id", repo.GetID()),
			attribute.Int("pull_request.number", pr.GetNumber()),
			attribute.String("pull_request.head_sha", pr.GetHead().GetSHA()),
		),
	)
	defer span.End()

	client, err := h.Client(ctx)
	if err != nil {
		return nil, err
	}

	list, _, err := client.Checks.ListCheckRunsForRef(ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		fmt.Sprintf("pull/%d/head", pr.GetNumber()),
		&github.ListCheckRunsOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	for _, v := range list.CheckRuns {
		checks = append(checks, Check{
			ID:         v.GetID(),
			Name:       v.GetName(),
			Status:     v.GetStatus(),
			Conclusion: v.GetConclusion(),
		})
	}

	status := generateChecksStatus(checks)
	span.AddEvent("Got checks", trace.WithAttributes(
		attribute.String("checks.status", status),
	))

	return checks, nil
}

func (h *Webhook) upsertCheck(ctx context.Context, c *github.CheckRunEvent) (pr *github.PullRequest, _ error) {
	ctx, span := h.tracer.Start(ctx, "UpsertCheck",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	run := c.GetCheckRun()
	for _, pr = range run.PullRequests {
		break
	}
	if pr == nil {
		span.AddEvent("NoPullRequestID")
		return nil, nil
	}

	tx, err := h.db.Tx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "begin tx")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := tx.Check.Create().
		SetID(run.GetCheckSuite().GetID()).
		SetName(run.GetName()).
		SetStatus(run.GetStatus()).
		SetConclusion(run.GetConclusion()).
		SetRepoID(c.GetRepo().GetID()).
		SetPullRequestID(pr.GetNumber()).
		OnConflict(
			sql.ConflictColumns(check.FieldID),
			sql.ResolveWithNewValues(),
		).
		UpdateNewValues().
		Exec(ctx); err != nil {
		return nil, errors.Wrap(err, "upsert check")
	}

	if err := h.fillPRState(ctx, tx.PRNotification, c.GetRepo(), pr); err != nil && !ent.IsNotFound(err) {
		return nil, errors.Wrap(err, "get PR state")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit")
	}

	return pr, nil
}
