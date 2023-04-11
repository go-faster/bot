package dispatch

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
)

func (b *Bot) OnBotCallbackQuery(ctx context.Context, e tg.Entities, u *tg.UpdateBotCallbackQuery) error {
	b.logger.Info("Got callback query")
	ctx, span := b.tracer.Start(ctx, "OnBotCallbackQuery")
	defer span.End()

	user, ok := e.Users[u.UserID]
	if !ok {
		return errors.Errorf("unknown user ID %d", u.UserID)
	}

	if err := b.onButton.OnButton(ctx, Button{
		QueryID: u.QueryID,
		User:    user.AsInput(),

		user:      user,
		baseEvent: b.baseEvent(),
	}); err != nil {
		return errors.Wrap(err, "handle onButton")
	}

	return nil
}
