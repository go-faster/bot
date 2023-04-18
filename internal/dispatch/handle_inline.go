package dispatch

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

func (b *Bot) OnBotInlineQuery(ctx context.Context, e tg.Entities, u *tg.UpdateBotInlineQuery) error {
	ctx, span := b.tracer.Start(ctx, "OnBotInlineQuery")
	defer span.End()

	zctx.From(ctx).Info("Got inline query",
		zap.String("query", u.Query),
		zap.String("offset", u.Offset),
	)

	user, ok := e.Users[u.UserID]
	if !ok {
		return errors.Errorf("unknown user ID %d", u.UserID)
	}

	var geo *tg.GeoPoint
	if u.Geo != nil {
		geo, _ = u.Geo.AsNotEmpty()
	}
	if err := b.onInline.OnInline(ctx, InlineQuery{
		QueryID:   u.QueryID,
		Query:     u.Query,
		Offset:    u.Offset,
		Enquirer:  user.AsInput(),
		geo:       geo,
		user:      user,
		baseEvent: b.baseEvent(ctx),
	}); err != nil {
		return errors.Wrap(err, "handle inline")
	}

	return nil
}
