package gh

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/ent/check"
	"github.com/go-faster/bot/internal/ent/lastchannelmessage"
	"github.com/go-faster/bot/internal/ent/prnotification"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/multierr"
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
			sql.ConflictColumns("id"),
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

func (h *Webhook) findPRNotification(ctx context.Context, channelID int64, pr *github.PullRequestEvent) (msgID, lastMsgID int, rerr error) {
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
			return 0, 0, errors.Wrap(err, "query pr notification")
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

func (h *Webhook) setPRNotification(ctx context.Context, pr *github.PullRequestEvent, msgID int) error {
	ctx, span := h.tracer.Start(ctx, "SetPRNotification",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	if err := h.db.PRNotification.Create().
		SetPullRequestID(pr.GetPullRequest().GetNumber()).
		SetRepoID(pr.GetRepo().GetID()).
		SetMessageID(msgID).
		OnConflict(
			sql.ConflictColumns("repo_id", "pull_request_id"),
			sql.ResolveWithNewValues(),
		).
		UpdateNewValues().
		Exec(ctx); err != nil {
		return errors.Wrap(err, "upsert PR notification")
	}

	return nil
}

type Check struct {
	ID         int64
	Name       string // e.g. "build"
	Conclusion string // failure, success, neutral, cancelled, skipped, timed_out, action_required
	Status     string // completed
}

func (h *Webhook) upsertCheck(ctx context.Context, c *github.CheckRunEvent) ([]Check, error) {
	ctx, span := h.tracer.Start(ctx, "UpsertCheck",
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
		SetPullRequestID(pullRequestID).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).
		UpdateNewValues().
		Exec(ctx); err != nil {
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
