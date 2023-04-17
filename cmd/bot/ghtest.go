package main

import (
	"context"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"

	"github.com/go-faster/bot/internal/dispatch"
)

func (a *App) HandleGitHubTest(ctx context.Context, e dispatch.MessageEvent) error {
	tok, _, err := a.github.Apps.CreateInstallationToken(ctx, 26766968, &github.InstallationTokenOptions{})
	if err != nil {
		if _, err := e.Reply().Textf(ctx, "Error: %v", err); err != nil {
			return err
		}
		return nil
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tok.GetToken()},
	)
	tc := oauth2.NewClient(ctx, ts)
	ghc := github.NewClient(tc)
	repo, _, err := ghc.Repositories.Get(ctx, "go-faster", "bot")
	if err != nil {
		if _, err := e.Reply().Textf(ctx, "Error: %v", err); err != nil {
			return err
		}
	}
	if _, err := e.Reply().Textf(ctx, "Repo id: %d", repo.GetID()); err != nil {
		return err
	}
	return nil
}
