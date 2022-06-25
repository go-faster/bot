package main

import (
	"context"

	"github.com/go-faster/bot/internal/dispatch"
)

func setupBot(a *App) error {
	a.mux.HandleFunc("/bot", "Ping bot", func(ctx context.Context, e dispatch.MessageEvent) error {
		_, err := e.Reply().Text(ctx, "What?")
		return err
	})
	a.mux.Handle("/stat", "Metrics and version", a.m.NewHandler())
	return nil
}
