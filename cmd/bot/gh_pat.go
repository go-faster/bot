package main

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/errors"

	"github.com/go-faster/bot/internal/dispatch"
)

func (a *App) HandleGitHubPersonalToken(ctx context.Context, e dispatch.MessageEvent) error {
	u, ok := e.User()
	if !ok {
		if _, err := e.Reply().Text(ctx, "Only for users"); err != nil {
			return errors.Wrap(err, "reply")
		}
		return nil
	}

	tok := e.Message.Message
	tok = strings.TrimPrefix(tok, "/gh_pat")
	tok = strings.TrimSpace(tok)

	if len(tok) == 0 {
		if _, err := e.Reply().Text(ctx, "Please, provide GitHub personal token"); err != nil {
			return errors.Wrap(err, "reply")
		}
		return nil
	}

	if err := a.db.User.Create().
		SetID(u.ID).
		SetUsername(u.Username).
		SetFirstName(u.FirstName).
		SetLastName(u.LastName).
		SetGithubToken(tok).OnConflict(
		sql.ConflictColumns("id"),
		sql.ResolveWithNewValues(),
	).UpdateGithubToken().Exec(ctx); err != nil {
		if _, err := e.Reply().Text(ctx, "500: Failed\nSorry, internal server error."); err != nil {
			return errors.Wrap(err, "reply")
		}
	}

	if _, err := e.Reply().Text(ctx, "✔️ Token set up"); err != nil {
		return errors.Wrap(err, "reply")
	}

	return nil
}
