package gh

import (
	"context"
	"sync"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

type updater struct {
	w *Webhook

	tick       time.Duration
	updates    map[prKey]PullRequestUpdate
	updatesMux sync.Mutex
}

func newUpdater(w *Webhook, tick time.Duration) *updater {
	return &updater{
		w:    w,
		tick: tick,
		// TODO(tdakkoa): store queue in DB?
		updates: map[prKey]PullRequestUpdate{},
	}
}

func (u *updater) updateOne(ctx context.Context, update PullRequestUpdate) error {
	// Do not query checks if PR was merged: we won't send status anyway.
	if update.Action != "merged" && update.Checks == nil {
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

	lg := zctx.From(ctx)

	for key, update := range u.updates {
		if err := u.updateOne(ctx, update); err != nil {
			lg.Error("PR Update failed", zap.Inline(key), zap.Error(err))
			continue
		}
		delete(u.updates, key)
	}
}

// Emit enqueues update.
func (u *updater) Emit(update PullRequestUpdate) {
	u.updatesMux.Lock()
	defer u.updatesMux.Unlock()

	key := prKey{
		update.Repo.GetFullName(),
		update.PR.GetNumber(),
	}
	u.updates[key] = update
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
