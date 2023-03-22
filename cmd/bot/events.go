package main

import (
	"context"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	"github.com/go-faster/errors"

	"github.com/go-faster/bot/internal/dispatch"
)

func (a *App) HandleEvents(ctx context.Context, e dispatch.MessageEvent) error {
	var count proto.ColUInt64
	q := ch.Query{
		Body: "SELECT COUNT() FROM faster.github_events_raw WHERE ts > now() - toIntervalMinute(30)",
		Result: proto.Results{
			{Name: "count()", Data: &count},
		},
	}
	if err := a.ch.Do(ctx, q); err != nil {
		return errors.Wrap(err, "do")
	}
	if len(count) == 0 {
		return errors.New("no data")
	}
	if _, err := e.Reply().Textf(ctx, "Events in last 30 minutes: %d", count[0]); err != nil {
		return errors.Wrap(err, "reply")
	}

	return nil
}
