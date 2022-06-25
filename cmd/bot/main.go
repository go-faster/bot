package main

import (
	"context"
	"os"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/go-faster/bot/internal/app"
)

func main() {
	app.Run(func(ctx context.Context, lg *zap.Logger) error {
		m, err := app.NewMetrics(lg, app.Config{
			Name:      "bot",
			Namespace: "faster",
			Addr:      os.Getenv("METRICS_ADDR"),
		})
		if err != nil {
			return errors.Wrap(err, "metrics")
		}
		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			return runBot(ctx, m, lg.Named("bot"))
		})
		g.Go(func() error {
			return m.Run(ctx)
		})

		return g.Wait()
	})
}
