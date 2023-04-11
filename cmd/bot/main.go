package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/app"
)

func main() {
	app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) error {
		return runBot(ctx, m, lg.Named("bot"))
	})
}
