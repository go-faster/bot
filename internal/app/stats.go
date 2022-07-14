package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gotd/td/tg"

	"github.com/go-faster/bot/internal/dispatch"
)

// Handler implements stats request handler.
type Handler struct {
	m *Metrics
}

func (m *Metrics) NewHandler() Handler {
	return Handler{m: m}
}

func (h Handler) stats() string {
	var w strings.Builder
	fmt.Fprintf(&w, "Statistics:\n\n")
	fmt.Fprintln(&w, "Messages:", h.m.Messages.Load())
	fmt.Fprintln(&w, "Responses:", h.m.Responses.Load())
	fmt.Fprintln(&w, "Media:", humanize.IBytes(uint64(h.m.MediaBytes.Load())))
	fmt.Fprintln(&w, "Uptime:", time.Since(h.m.Start).Round(time.Second))
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
