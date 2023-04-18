package dispatch

import (
	"context"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

type handle struct {
	MessageHandler
	description string
}

// MessageMux is message event router.
type MessageMux struct {
	prefixes map[string]handle
	fallback MessageHandler
	tracer   trace.Tracer
}

// NewMessageMux creates new MessageMux.
func NewMessageMux() *MessageMux {
	return &MessageMux{prefixes: map[string]handle{}, tracer: trace.NewNoopTracerProvider().Tracer("")}
}

func (m *MessageMux) WithTracerProvider(provider trace.TracerProvider) *MessageMux {
	m.tracer = provider.Tracer("td.dispatch.message_mux")
	return m
}

// Handle adds given prefix and handler to the mux.
func (m *MessageMux) Handle(prefix, description string, handler MessageHandler) {
	m.prefixes[prefix] = handle{
		MessageHandler: handler,
		description:    description,
	}
}

// HandleFunc adds given prefix and handler to the mux.
func (m *MessageMux) HandleFunc(prefix, description string, handler func(ctx context.Context, e MessageEvent) error) {
	m.Handle(prefix, description, MessageHandlerFunc(handler))
}

// OnMessage implements MessageHandler.
func (m *MessageMux) OnMessage(ctx context.Context, e MessageEvent) error {
	ctx, span := m.tracer.Start(ctx, "OnMessage")
	defer span.End()

	lg := zctx.From(ctx)

	for prefix, handler := range m.prefixes {
		if strings.HasPrefix(e.Message.Message, prefix) {
			lg.Debug("Found handler", zap.String("prefix", prefix))
			if err := handler.OnMessage(ctx, e); err != nil {
				return errors.Wrapf(err, "handle %q", prefix)
			}
			return nil
		}
	}

	if h := m.fallback; h != nil {
		lg.Debug("Using fallback")
		return h.OnMessage(ctx, e)
	}

	lg.Debug("No handler found")

	return nil
}

// SetFallback sets fallback handler, if mux is unable to find a command handler.
func (m *MessageMux) SetFallback(h MessageHandler) {
	m.fallback = h
}

// SetFallbackFunc sets fallback handler, if mux is unable to find a command handler.
func (m *MessageMux) SetFallbackFunc(h func(ctx context.Context, e MessageEvent) error) {
	m.SetFallback(MessageHandlerFunc(h))
}

// RegisterCommands registers all mux commands using https://core.telegram.org/method/bots.setBotCommands.
func (m *MessageMux) RegisterCommands(ctx context.Context, raw *tg.Client) error {
	commands := make([]tg.BotCommand, 0, len(m.prefixes))
	for prefix, handler := range m.prefixes {
		if handler.description == "" {
			continue
		}
		commands = append(commands, tg.BotCommand{
			Command:     strings.TrimPrefix(prefix, "/"),
			Description: handler.description,
		})
	}

	if _, err := raw.BotsSetBotCommands(ctx, &tg.BotsSetBotCommandsRequest{
		Scope:    &tg.BotCommandScopeDefault{},
		LangCode: "en",
		Commands: commands,
	}); err != nil {
		return errors.Wrap(err, "set commands")
	}
	return nil
}
