package dispatch

import (
	"context"

	"github.com/gotd/td/tg"
)

type Button struct {
	QueryID int64
	User    *tg.InputUser

	user *tg.User

	baseEvent
}

type ButtonHandler interface {
	OnButton(ctx context.Context, e Button) error
}

// ButtonHandlerFunc is a functional adapter for Handler.
type ButtonHandlerFunc func(ctx context.Context, e Button) error

// OnButton implements ButtonHandler.
func (h ButtonHandlerFunc) OnButton(ctx context.Context, e Button) error {
	return h(ctx, e)
}
