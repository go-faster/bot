package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/dispatch"
)

func (a *App) OnButton(ctx context.Context, e dispatch.Button) error {
	if e.User != nil {
		a.lg.Info("OnButton",
			zap.String("user", e.User.String()),
		)
	} else {
		a.lg.Info("OnButton: no user")
	}
	return nil
}
