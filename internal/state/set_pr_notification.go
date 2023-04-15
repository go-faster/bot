package state

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/trace"
)

func (e Ent) SetPRNotification(ctx context.Context, pr *github.PullRequestEvent, msgID int) error {
	ctx, span := e.tracer.Start(ctx, "SetPRNotification",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	if err := e.db.PRNotification.Create().
		SetPullRequestID(pr.GetPullRequest().GetNumber()).
		SetRepoID(pr.GetRepo().GetID()).
		SetMessageID(msgID).
		OnConflict(
			sql.ConflictColumns("repo_id", "pull_request_id"),
			sql.ResolveWithNewValues(),
		).UpdateNewValues().Exec(ctx); err != nil {
		return errors.Wrap(err, "upsert")
	}
	return nil
}
