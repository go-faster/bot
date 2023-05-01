package gh

import (
	"context"
	"sync"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-faster/bot/internal/ent"
)

type prKey struct {
	// repo is full repo name.
	repo string
	// number is pr number.
	number int
}

func (k prKey) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("repo", k.repo)
	e.AddInt("pr", k.number)
	return nil
}

type queuedUpdate struct {
	Update PullRequestUpdate
	Tries  int
}

type updater struct {
	w *Webhook

	tick         time.Duration
	prUpdates    map[prKey]*queuedUpdate
	checkUpdates map[prKey]*queuedUpdate
	updatesMux   sync.Mutex
}

func newUpdater(w *Webhook, tick time.Duration) *updater {
	return &updater{
		w:    w,
		tick: tick,
		// TODO(tdakkota): store queue in DB?
		prUpdates:    map[prKey]*queuedUpdate{},
		checkUpdates: map[prKey]*queuedUpdate{},
	}
}

var errNoNotificationYet = errors.New("no PR notification message yet")

func (u *updater) updateOne(ctx context.Context, update PullRequestUpdate) error {
	if update.Event == "check_update" {
		switch err := u.w.fillPRState(ctx, u.w.db.PRNotification, update.Repo, update.PR); {
		case err == nil:
		case ent.IsNotFound(err):
			return errNoNotificationYet
		default:
			return errors.Wrap(err, "query cached pr fields")
		}
	}

	// Do not query checks if PR was merged: we won't send status anyway.
	if !update.ActionIn("merged", "closed") && update.Checks == nil {
		checks, err := u.w.queryChecks(ctx, update.Repo, update.PR)
		if err != nil {
			return errors.Wrap(err, "query checks")
		}
		update.Checks = checks
	}

	return u.w.updatePR(ctx, update)
}

func (u *updater) doUpdate(ctx context.Context) {
	u.updatesMux.Lock()
	defer u.updatesMux.Unlock()

	applyUpdates := func(updates map[prKey]*queuedUpdate) {
		for key, qu := range updates {
			ctx := zctx.With(ctx,
				zap.String("event", qu.Update.Event),
				zap.Inline(key),
			)

			if err := u.updateOne(ctx, qu.Update); err != nil {
				lg := zctx.From(ctx)
				if !errors.Is(err, errNoNotificationYet) {
					lg.Error("PR Update failed", zap.Error(err))
				} else {
					lg.Debug("Update checks later: no PR yet")
				}

				if qu.Tries < 5 {
					qu.Tries++
					continue
				}
			}
			delete(updates, key)
		}
	}
	applyUpdates(u.prUpdates)
	applyUpdates(u.checkUpdates)
}

// Emit enqueues update.
func (u *updater) Emit(update PullRequestUpdate) error {
	u.updatesMux.Lock()
	defer u.updatesMux.Unlock()

	key := prKey{
		update.Repo.GetFullName(),
		update.PR.GetNumber(),
	}
	emit := &queuedUpdate{
		Update: update,
	}

	switch e := update.Event; e {
	case "pr_update":
		u.prUpdates[key] = emit
		return nil
	case "check_update":
		u.checkUpdates[key] = emit
		return nil
	default:
		return errors.Errorf("unexpected event type %q", e)
	}
}

// Run setups update worker.
func (u *updater) Run(ctx context.Context) error {
	t := time.NewTicker(u.tick)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			u.doUpdate(ctx)
		}
	}
}
