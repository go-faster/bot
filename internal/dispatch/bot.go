package dispatch

import (
	"context"
	"crypto/rand"
	"io"

	"go.uber.org/zap"

	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

// Bot represents generic Telegram bot state and event dispatcher.
type Bot struct {
	onMessage MessageHandler
	onInline  InlineHandler

	rpc        *tg.Client
	sender     *message.Sender
	downloader *downloader.Downloader

	logger *zap.Logger
	rand   io.Reader
}

// NewBot creates new bot.
func NewBot(raw *tg.Client) *Bot {
	return &Bot{
		onMessage: MessageHandlerFunc(func(context.Context, MessageEvent) error {
			return nil
		}),
		onInline: InlineHandlerFunc(func(context.Context, InlineQuery) error {
			return nil
		}),
		rpc:        raw,
		sender:     message.NewSender(raw),
		downloader: downloader.NewDownloader(),
		logger:     zap.NewNop(),
		rand:       rand.Reader,
	}
}

// OnMessage sets message handler.
func (b *Bot) OnMessage(handler MessageHandler) *Bot {
	b.onMessage = handler
	return b
}

// OnInline sets inline query handler.
func (b *Bot) OnInline(handler InlineHandler) *Bot {
	b.onInline = handler
	return b
}

// WithSender sets message sender to use.
func (b *Bot) WithSender(sender *message.Sender) *Bot {
	b.sender = sender
	return b
}

// WithLogger sets logger.
func (b *Bot) WithLogger(logger *zap.Logger) *Bot {
	b.logger = logger
	return b
}

// Register sets handlers using given dispatcher.
func (b *Bot) Register(dispatcher tg.UpdateDispatcher) *Bot {
	dispatcher.OnNewMessage(b.OnNewMessage)
	dispatcher.OnNewChannelMessage(b.OnNewChannelMessage)
	dispatcher.OnBotInlineQuery(b.OnBotInlineQuery)
	return b
}
