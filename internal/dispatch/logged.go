package dispatch

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// LoggedDispatcher is update logging middleware.
type LoggedDispatcher struct {
	handler telegram.UpdateHandler
	log     *zap.Logger
}

// NewLoggedDispatcher creates new update logging middleware.
func NewLoggedDispatcher(next telegram.UpdateHandler, log *zap.Logger) LoggedDispatcher {
	return LoggedDispatcher{
		handler: next,
		log:     log,
	}
}

// Handle implements telegram.UpdateHandler.
func (d LoggedDispatcher) Handle(ctx context.Context, u tg.UpdatesClass) error {
	d.log.Debug("Update",
		zap.String("t", fmt.Sprintf("%T", u)),
	)
	return d.handler.Handle(ctx, u)
}
