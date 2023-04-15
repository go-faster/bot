package state

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-faster/bot/internal/ent/lastchannelmessage"
	"github.com/go-faster/bot/internal/ent/prnotification"
)

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
