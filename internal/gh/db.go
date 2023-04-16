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

	if err := h.db.PRNotification.Create().
		SetRepoID(repo.GetID()).
		SetPullRequestID(pr.GetNumber()).
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

type Check struct {
	ID         int64
	Name       string // e.g. "build"
	Conclusion string // failure, success, neutral, cancelled, skipped, timed_out, action_required
	Status     string // completed
}

func queryChecks(ctx context.Context, tx *ent.CheckClient, repo *github.Repository, pr *github.PullRequest) (checks []Check, _ error) {
	list, err := tx.Query().
		Where(
			check.RepoID(repo.GetID()),
			check.PullRequestIDEQ(pr.GetNumber()),
		).
		All(ctx)
	if err != nil {
		return checks, err
	}

	for _, v := range list {
		checks = append(checks, Check{
			ID:         v.ID,
			Name:       v.Name,
			Status:     v.Status,
			Conclusion: v.Conclusion,
		})
	}

	return checks, nil
}

func (h *Webhook) queryChecks(ctx context.Context, repo *github.Repository, pr *github.PullRequest) (checks []Check, _ error) {
	return queryChecks(ctx, h.db.Check, repo, pr)
}

func (h *Webhook) upsertCheck(ctx context.Context, c *github.CheckRunEvent) (pr *github.PullRequest, _ []Check, _ error) {
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
		return nil, nil, nil
	}

	tx, err := h.db.Tx(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "begin tx")
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
		return nil, nil, errors.Wrap(err, "upsert check")
	}

	checks, err := queryChecks(ctx, h.db.Check, c.GetRepo(), pr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "query checks")
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, errors.Wrap(err, "commit")
	}

	return pr, checks, nil
}
