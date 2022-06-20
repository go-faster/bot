package main

import (
	"context"

	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/metrics"
)

func setupBot(app *App) error {
	app.mux.HandleFunc("/bot", "Ping bot", func(ctx context.Context, e dispatch.MessageEvent) error {
		_, err := e.Reply().Text(ctx, "What?")
		return err
	})
	app.mux.Handle("/stat", "Metrics and version", metrics.NewHandler(app.mts))
	return nil
}
