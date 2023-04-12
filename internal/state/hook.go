package state

import (
	"context"

	"github.com/google/go-github/v50/github"
	"go.uber.org/multierr"

	"github.com/go-faster/bot/internal/dispatch"
)

// Hook is event handler which saves last message ID of dialog to the Pebble storage.
type Hook struct {
	next    dispatch.MessageHandler
	storage Storage
}

type Storage interface {
	UpdateLastMsgID(ctx context.Context, channelID int64, msgID int) error
	SetPRNotification(ctx context.Context, pr *github.PullRequestEvent, msgID int) error
	FindPRNotification(ctx context.Context, channelID int64, pr *github.PullRequestEvent) (msgID, lastMsgID int, rerr error)
}

// NewHook creates new hook.
func NewHook(next dispatch.MessageHandler, storage Storage) Hook {
	return Hook{next: next, storage: storage}
}

// OnMessage implements dispatch.MessageHandler.
func (h Hook) OnMessage(ctx context.Context, e dispatch.MessageEvent) error {
	ch, ok := e.Channel()
	if !ok {
		return h.next.OnMessage(ctx, e)
	}

	return multierr.Append(
		h.storage.UpdateLastMsgID(context.Background(), ch.ID, e.Message.ID),
		h.next.OnMessage(ctx, e),
	)
}
