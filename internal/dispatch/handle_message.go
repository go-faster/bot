package dispatch

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/go-faster/simon/sdk/zctx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

func (b *Bot) handleUser(ctx context.Context, user, sender *tg.User, m *tg.Message) error {
	zctx.From(ctx).Info("Got message",
		zap.String("text", m.Message),
		zap.Int64("user_id", user.ID),
		zap.String("user_first_name", user.FirstName),
		zap.String("username", user.Username),
	)

	return b.onMessage.OnMessage(ctx, MessageEvent{
		Peer:    user.AsInputPeer(),
		Message: m,

		msgSender: sender,
		user:      user,
		baseEvent: b.baseEvent(ctx),
	})
}

func (b *Bot) handleChat(ctx context.Context, chat *tg.Chat, sender *tg.User, m *tg.Message) error {
	zctx.From(ctx).Info("Got message from chat",
		zap.String("text", m.Message),
		zap.Int64("chat_id", chat.ID),
	)

	return b.onMessage.OnMessage(ctx, MessageEvent{
		Peer:    chat.AsInputPeer(),
		Message: m,

		msgSender: sender,
		chat:      chat,
		baseEvent: b.baseEvent(ctx),
	})
}

func (b *Bot) handleChannel(ctx context.Context, channel *tg.Channel, sender *tg.User, m *tg.Message) error {
	zctx.From(ctx).Info("Got message from channel",
		zap.String("text", m.Message),
		zap.String("username", channel.Username),
		zap.Int64("channel_id", channel.ID),
	)

	return b.onMessage.OnMessage(ctx, MessageEvent{
		Peer:    channel.AsInputPeer(),
		Message: m,

		msgSender: sender,
		channel:   channel,
		baseEvent: b.baseEvent(ctx),
	})
}

func (b *Bot) handleMessage(ctx context.Context, e tg.Entities, msg tg.MessageClass) error {
	ctx, span := b.tracer.Start(ctx, "handleMessage")
	defer span.End()

	switch m := msg.(type) {
	case *tg.Message:
		if m.Out {
			return nil
		}

		var sender *tg.User
		if p, ok := m.GetFromID(); ok {
			lg := zctx.From(ctx).With(zap.String("from_peer", fmt.Sprintf("%#v", p)))
			zctx.With(ctx, lg)

			pu, ok := p.(*tg.PeerUser)
			if !ok {
				lg.Info("Not gonna answer to non-user sender")
				return nil
			}
			sender = e.Users[pu.UserID]

			if sender != nil && sender.Bot {
				lg.Info("Not gonna answer to bot")
				return nil
			}
		}

		switch p := m.PeerID.(type) {
		case *tg.PeerUser:
			user, ok := e.Users[p.UserID]
			if !ok {
				return errors.Errorf("unknown user ID %d", p.UserID)
			}

			if sender == nil {
				sender = user
			}

			return b.handleUser(ctx, user, sender, m)
		case *tg.PeerChat:
			chat, ok := e.Chats[p.ChatID]
			if !ok {
				return errors.Errorf("unknown chat ID %d", p.ChatID)
			}
			return b.handleChat(ctx, chat, sender, m)
		case *tg.PeerChannel:
			channel, ok := e.Channels[p.ChannelID]
			if !ok {
				return errors.Errorf("unknown channel ID %d", p.ChannelID)
			}
			return b.handleChannel(ctx, channel, sender, m)
		}
	}

	return nil
}

func (b *Bot) logError(ctx context.Context, span trace.Span, msg tg.MessageClass, rerr *error) {
	if *rerr == nil {
		return
	}
	zctx.From(ctx).Error("Message handler error",
		zap.Int("msg_id", msg.GetID()),
		zap.Error(*rerr),
	)
	span.SetAttributes(
		attribute.Int("telegram.message_id", msg.GetID()),
	)
	span.RecordError(*rerr)
}

func (b *Bot) OnNewMessage(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) (rerr error) {
	ctx, span := b.tracer.Start(ctx, "OnNewMessage")
	defer span.End()

	defer b.logError(ctx, span, u.Message, &rerr)

	if err := b.handleMessage(ctx, e, u.Message); err != nil {
		if !tg.IsUserBlocked(err) {
			return errors.Wrapf(err, "handle message %d", u.Message.GetID())
		}

		zctx.From(ctx).Debug("Bot is blocked by user")
	}
	return nil
}

func (b *Bot) OnNewChannelMessage(ctx context.Context, e tg.Entities, u *tg.UpdateNewChannelMessage) (rerr error) {
	ctx, span := b.tracer.Start(ctx, "OnNewChannelMessage")
	defer span.End()
	defer b.logError(ctx, span, u.Message, &rerr)

	if err := b.handleMessage(ctx, e, u.Message); err != nil {
		return errors.Wrap(err, "handle")
	}
	return nil
}
