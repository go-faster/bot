package main

import (
	"context"

	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/gpt"
	"github.com/go-faster/errors"
)

func setupBot(a *App) error {
	a.mux.HandleFunc("/bot", "Ping bot", func(ctx context.Context, e dispatch.MessageEvent) error {
		_, err := e.Reply().Text(ctx, "What?")
		return err
	})
	a.mux.Handle("/stat", "Metrics and version", a.m.NewHandler())
	a.mux.HandleFunc("/events", "GitHub events", a.HandleEvents)
	a.mux.HandleFunc("/gh_pat", "Set GitHub personal token", a.HandleGitHubPersonalToken)
	{
		var limitCfg gpt.LimitConfig
		if err := limitCfg.ParseEnv(); err != nil {
			return errors.Wrap(err, "parse GPT limit config")
		}
		hgpt := gpt.New(a.openai, a.db, a.m.TracerProvider()).
			WithLimitConfig(limitCfg)
		a.mux.HandleFunc("/gpt", "ChatGPT 3.5", hgpt.OnCommand)
		a.mux.SetFallbackFunc(hgpt.OnReply)
	}
	return nil
}
