package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-faster/sdk/app"
	"github.com/gotd/contrib/oteltg"
	"github.com/gotd/td/tg"

	"github.com/go-faster/bot/internal/dispatch"
)

// Handler implements stats request handler.
type Handler struct {
	Middleware *oteltg.Middleware

	m *app.Metrics
}

func NewHandler(m *app.Metrics) Handler {
	return Handler{m: m}
}

func (h Handler) stats() string {
	var w strings.Builder
	fmt.Fprintf(&w, "Statistics:\n\n")
	fmt.Fprintln(&w, "TL Layer version:", tg.Layer)
	if v := GetGotdVersion(); v != "" {
		fmt.Fprintln(&w, "Version:", v)
	}

	return w.String()
}

// OnMessage implements dispatch.MessageHandler.
func (h Handler) OnMessage(ctx context.Context, e dispatch.MessageEvent) error {
	_, err := e.Reply().Text(ctx, h.stats())
	return err
}
