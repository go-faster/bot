package dispatch

import (
	"context"
	"crypto/rand"
	"io"

	"github.com/google/go-github/v52/github"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"go.opentelemetry.io/otel/trace"
)

// Bot represents generic Telegram bot state and event dispatcher.
type Bot struct {
	onMessage MessageHandler
	onInline  InlineHandler
	onButton  ButtonHandler

	rpc        *tg.Client
	sender     *message.Sender
	downloader *downloader.Downloader
	github     *github.Client

	rand   io.Reader
	tracer trace.Tracer
}

const botInstrumentationName = "bot"

// NewBot creates new bot.
func NewBot(raw *tg.Client) *Bot {
	return &Bot{
		onMessage: MessageHandlerFunc(func(context.Context, MessageEvent) error {
			return nil
		}),
		onInline: InlineHandlerFunc(func(context.Context, InlineQuery) error {
			return nil
		}),
		onButton: ButtonHandlerFunc(func(context.Context, Button) error {
			return nil
		}),
		rpc:        raw,
		sender:     message.NewSender(raw),
		downloader: downloader.NewDownloader(),
		rand:       rand.Reader,
		tracer:     trace.NewNoopTracerProvider().Tracer(botInstrumentationName),
	}
}

// OnMessage sets message handler.
func (b *Bot) OnMessage(handler MessageHandler) *Bot {
	b.onMessage = handler
	return b
}

func (b *Bot) WithGitHub(client *github.Client) *Bot {
	b.github = client
	return b
}

// OnInline sets inline query handler.
func (b *Bot) OnInline(handler InlineHandler) *Bot {
	b.onInline = handler
	return b
}

// OnButton sets button handler.
func (b *Bot) OnButton(handler ButtonHandler) *Bot {
	b.onButton = handler
	return b
}

// WithSender sets message sender to use.
func (b *Bot) WithSender(sender *message.Sender) *Bot {
	b.sender = sender
	return b
}

func (b *Bot) WithTracerProvider(provider trace.TracerProvider) *Bot {
	b.tracer = provider.Tracer(botInstrumentationName)
	return b
}

// Register sets handlers using given dispatcher.
func (b *Bot) Register(dispatcher tg.UpdateDispatcher) *Bot {
	dispatcher.OnNewMessage(b.OnNewMessage)
	dispatcher.OnNewChannelMessage(b.OnNewChannelMessage)
	dispatcher.OnBotInlineQuery(b.OnBotInlineQuery)
	dispatcher.OnBotCallbackQuery(b.OnBotCallbackQuery)
	return b
}
