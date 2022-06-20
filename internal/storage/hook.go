package storage

import (
	"context"

	"go.uber.org/multierr"

	"github.com/go-faster/bot/internal/dispatch"
)

// Hook is event handler which saves last message ID of dialog to the Pebble storage.
type Hook struct {
	next    dispatch.MessageHandler
	storage MsgID
}

// NewHook creates new hook.
func NewHook(next dispatch.MessageHandler, storage MsgID) Hook {
	return Hook{next: next, storage: storage}
}

// OnMessage implements dispatch.MessageHandler.
func (h Hook) OnMessage(ctx context.Context, e dispatch.MessageEvent) error {
	ch, ok := e.Channel()
	if !ok {
		return h.next.OnMessage(ctx, e)
	}

	return multierr.Append(
		h.storage.UpdateLastMsgID(ch.ID, e.Message.ID),
		h.next.OnMessage(ctx, e),
	)
}
