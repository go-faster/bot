package main

import (
	"context"

	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/gpt"
)

func setupBot(a *App) error {
	a.mux.HandleFunc("/bot", "Ping bot", func(ctx context.Context, e dispatch.MessageEvent) error {
		_, err := e.Reply().Text(ctx, "What?")
		return err
	})
	a.mux.Handle("/gpt", "ChatGPT 3.5", gpt.New(a.openai))
	a.mux.Handle("/stat", "Metrics and version", a.m.NewHandler())
	a.mux.HandleFunc("/events", "GitHub events", a.HandleEvents)
	return nil
}
