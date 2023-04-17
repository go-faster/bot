package main

import (
	"context"

	"github.com/go-faster/bot/internal/dispatch"
)

func (a *App) HandleGitHubTest(ctx context.Context, e dispatch.MessageEvent) error {
	repo, _, err := a.github.Repositories.Get(ctx, "go-faster", "bot")
	if err != nil {
		if _, err := e.Reply().Textf(ctx, "Error: %+v", err); err != nil {
			return err
		}
		return nil
	}
	if _, err := e.Reply().Textf(ctx, "Repo id: %d", repo.GetID()); err != nil {
		return err
	}
	return nil
}
