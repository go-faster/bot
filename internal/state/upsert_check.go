package state

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-faster/bot/internal/ent/check"
)

func (e Ent) UpsertCheck(ctx context.Context, c *github.CheckRunEvent) ([]Check, error) {
	ctx, span := e.tracer.Start(ctx, "UpsertCheck",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	run := c.GetCheckRun()
	var pullRequestID int
	for _, pr := range run.PullRequests {
		pullRequestID = pr.GetNumber()
		break
	}
	if pullRequestID == 0 {
		span.AddEvent("NoPullRequestID")
		return nil, nil
	}

	tx, err := e.db.Tx(ctx)
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
		SetPullRequestID(pullRequestID).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateNewValues().Exec(ctx); err != nil {
		return nil, errors.Wrap(err, "upsert check")
	}

	var out []Check
	list, err := tx.Check.Query().Where(
		check.ID(run.GetID()),
		check.RepoID(c.GetRepo().GetID()),
		check.PullRequestIDEQ(pullRequestID),
	).All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "query checks")
	}
	for _, v := range list {
		out = append(out, Check{
			ID:         v.ID,
			Name:       v.Name,
			Status:     v.Status,
			Conclusion: v.Conclusion,
		})
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit")
	}
	return out, nil
}
