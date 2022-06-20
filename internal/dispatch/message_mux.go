package dispatch

import (
	"context"
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

type handle struct {
	MessageHandler
	description string
}

// MessageMux is message event router.
type MessageMux struct {
	prefixes map[string]handle
}

// NewMessageMux creates new MessageMux.
func NewMessageMux() MessageMux {
	return MessageMux{prefixes: map[string]handle{}}
}

// Handle adds given prefix and handler to the mux.
func (m MessageMux) Handle(prefix, description string, handler MessageHandler) {
	m.prefixes[prefix] = handle{
		MessageHandler: handler,
		description:    description,
	}
}

// HandleFunc adds given prefix and handler to the mux.
func (m MessageMux) HandleFunc(prefix, description string, handler func(ctx context.Context, e MessageEvent) error) {
	m.Handle(prefix, description, MessageHandlerFunc(handler))
}

// OnMessage implements MessageHandler.
func (m MessageMux) OnMessage(ctx context.Context, e MessageEvent) error {
	for prefix, handler := range m.prefixes {
		if strings.HasPrefix(e.Message.Message, prefix) {
			if err := handler.OnMessage(ctx, e); err != nil {
				return errors.Wrapf(err, "handle %q", prefix)
			}
			return nil
		}
	}

	return nil
}

// RegisterCommands registers all mux commands using https://core.telegram.org/method/bots.setBotCommands.
func (m MessageMux) RegisterCommands(ctx context.Context, raw *tg.Client) error {
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
