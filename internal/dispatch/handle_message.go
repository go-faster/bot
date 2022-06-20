package dispatch

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

func (b *Bot) handleUser(ctx context.Context, user *tg.User, m *tg.Message) error {
	b.logger.Info("Got message",
		zap.String("text", m.Message),
		zap.Int64("user_id", user.ID),
		zap.String("user_first_name", user.FirstName),
		zap.String("username", user.Username),
	)

	return b.onMessage.OnMessage(ctx, MessageEvent{
		Peer:      user.AsInputPeer(),
		user:      user,
		Message:   m,
		baseEvent: b.baseEvent(),
	})
}

func (b *Bot) handleChat(ctx context.Context, chat *tg.Chat, m *tg.Message) error {
	b.logger.Info("Got message from chat",
		zap.String("text", m.Message),
		zap.Int64("chat_id", chat.ID),
	)

	return b.onMessage.OnMessage(ctx, MessageEvent{
		Peer:      chat.AsInputPeer(),
		chat:      chat,
		Message:   m,
		baseEvent: b.baseEvent(),
	})
}

func (b *Bot) handleChannel(ctx context.Context, channel *tg.Channel, m *tg.Message) error {
	b.logger.Info("Got message from channel",
		zap.String("text", m.Message),
		zap.String("username", channel.Username),
		zap.Int64("channel_id", channel.ID),
	)

	return b.onMessage.OnMessage(ctx, MessageEvent{
		Peer:      channel.AsInputPeer(),
		channel:   channel,
		Message:   m,
		baseEvent: b.baseEvent(),
	})
}

func (b *Bot) handleMessage(ctx context.Context, e tg.Entities, msg tg.MessageClass) error {
	switch m := msg.(type) {
	case *tg.Message:
		if m.Out {
			return nil
		}

		switch p := m.PeerID.(type) {
		case *tg.PeerUser:
			user, ok := e.Users[p.UserID]
			if !ok {
				return errors.Errorf("unknown user ID %d", p.UserID)
			}
			return b.handleUser(ctx, user, m)
		case *tg.PeerChat:
			chat, ok := e.Chats[p.ChatID]
			if !ok {
				return errors.Errorf("unknown chat ID %d", p.ChatID)
			}
			return b.handleChat(ctx, chat, m)
		case *tg.PeerChannel:
			channel, ok := e.Channels[p.ChannelID]
			if !ok {
				return errors.Errorf("unknown channel ID %d", p.ChannelID)
			}
			return b.handleChannel(ctx, channel, m)
		}
	}

	return nil
}

func (b *Bot) OnNewMessage(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
	if err := b.handleMessage(ctx, e, u.Message); err != nil {
		if !tg.IsUserBlocked(err) {
			return errors.Wrapf(err, "handle message %d", u.Message.GetID())
		}

		b.logger.Debug("Bot is blocked by user")
	}
	return nil
}

func (b *Bot) OnNewChannelMessage(ctx context.Context, e tg.Entities, u *tg.UpdateNewChannelMessage) error {
	if err := b.handleMessage(ctx, e, u.Message); err != nil {
		return errors.Wrap(err, "handle")
	}
	return nil
}
