package main

import (
	"context"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	"github.com/go-faster/errors"

	"github.com/go-faster/bot/internal/dispatch"
)

func (a *App) HandleEvents(ctx context.Context, e dispatch.MessageEvent) error {
	var count proto.ColUInt64
	if err := a.ch.Do(ctx, ch.Query{
		Body: "SELECT COUNT() FROM faster.github_events_raw WHERE ts > now() - toIntervalMinute(30)",
		Result: proto.Results{
			{Name: "count()", Data: &count},
		},
	}); err != nil {
		return errors.Wrap(err, "do")
	}
	if len(count) == 0 {
		return errors.New("no data")
	}
	var max, min proto.ColDateTime
	if err := a.ch.Do(ctx, ch.Query{
		Body: "SELECT max(ts), min(ts) FROM faster.github_events_raw;",
		Result: proto.Results{
			{Name: "max(ts)", Data: &max},
			{Name: "min(ts)", Data: &min},
		},
	}); err != nil {
		return errors.Wrap(err, "do")
	}
	duration := max.Row(0).Sub(min.Row(0))
	if _, err := e.Reply().Textf(ctx, "Events in last 30 minutes: %d\nSize: %s",
		count[0], duration.Truncate(1*time.Second),
	); err != nil {
		return errors.Wrap(err, "reply")
	}

	return nil
}
