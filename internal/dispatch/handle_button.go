package dispatch

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/gotd/td/tg"
	"go.opentelemetry.io/otel/attribute"
)

func (b *Bot) OnBotCallbackQuery(ctx context.Context, e tg.Entities, u *tg.UpdateBotCallbackQuery) error {
	ctx, span := b.tracer.Start(ctx, "OnBotCallbackQuery")
	defer span.End()

	zctx.From(ctx).Info("Got callback query")

	user, ok := e.Users[u.UserID]
	if !ok {
		return errors.Errorf("unknown user ID %d", u.UserID)
	}

	span.SetAttributes(
		attribute.Int64("user.id", user.ID),
		attribute.String("user.username", user.Username),
		attribute.String("user.first_name", user.FirstName),
		attribute.String("user.last_name", user.LastName),
	)

	if err := b.onButton.OnButton(ctx, Button{
		QueryID: u.QueryID,
		Input:   user.AsInput(),
		Data:    u.Data,
		User:    user,

		baseEvent: b.baseEvent(ctx),
	}); err != nil {
		return errors.Wrap(err, "handle onButton")
	}

	return nil
}
