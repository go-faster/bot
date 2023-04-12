package state

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/ent/lastchannelmessage"
	"github.com/go-faster/bot/internal/ent/prnotification"
)

type Ent struct {
	db     *ent.Client
	tracer trace.Tracer
}

func (e Ent) UpdateLastMsgID(ctx context.Context, channelID int64, msgID int) error {
	ctx, span := e.tracer.Start(ctx, "UpdateLastMsgID",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	if err := e.db.LastChannelMessage.Create().
		SetID(channelID).
		SetMessageID(msgID).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).UpdateMessageID().
		Exec(ctx); err != nil {
		return errors.Wrap(err, "upsert")
	}
	return nil
}

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

func (e Ent) FindPRNotification(ctx context.Context, channelID int64, pr *github.PullRequestEvent) (msgID, lastMsgID int, rerr error) {
	ctx, span := e.tracer.Start(ctx, "FindPRNotification",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	tx, err := e.db.Tx(ctx)
	if err != nil {
		return 0, 0, errors.Wrap(err, "begin tx")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	{
		list, err := tx.LastChannelMessage.Query().Where(
			lastchannelmessage.IDEQ(channelID),
		).All(ctx)
		if err != nil {
			return 0, 0, errors.Wrap(err, "query last message")
		}
		for _, v := range list {
			lastMsgID = v.MessageID
		}
	}
	{
		list, err := tx.PRNotification.Query().Where(
			prnotification.PullRequestIDEQ(pr.GetPullRequest().GetNumber()),
			prnotification.RepoIDEQ(pr.GetRepo().GetID()),
		).All(ctx)
		if err != nil {
			return 0, 0, errors.Wrap(err, "query last message")
		}
		for _, v := range list {
			msgID = v.MessageID
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, errors.Wrap(err, "commit")
	}

	return msgID, lastMsgID, nil
}

func NewEnt(db *ent.Client, t trace.TracerProvider) *Ent {
	return &Ent{db: db, tracer: t.Tracer("state")}
}

var _ Storage = (*Ent)(nil)
