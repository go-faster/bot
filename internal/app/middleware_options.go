package app

import (
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/botapi"
)

// MiddlewareOptions is middleware options.
type MiddlewareOptions struct {
	BotAPI *botapi.Client
	Logger *zap.Logger
}

func (m *MiddlewareOptions) setDefaults() {
	if m.Logger == nil {
		m.Logger = zap.NewNop()
	}
}
