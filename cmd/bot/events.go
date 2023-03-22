package main

import (
	"context"
	"fmt"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	"github.com/dustin/go-humanize"
	"github.com/go-faster/errors"

	"github.com/go-faster/bot/internal/dispatch"
)

func formatInt(x int) string {
	return formatFloat(float64(x))
}

func formatFloat(num float64) string {
	v, u := humanize.ComputeSI(num)
	return fmt.Sprintf("%.1f%s", v, u)
}

func (a *App) HandleEvents(ctx context.Context, e dispatch.MessageEvent) error {
	var count, total proto.ColUInt64
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
	if err := a.ch.Do(ctx, ch.Query{
		Body: "SELECT COUNT() FROM faster.github_events_raw",
		Result: proto.Results{
			{Name: "count()", Data: &total},
		},
	}); err != nil {
		return errors.Wrap(err, "do")
	}
	if len(total) == 0 {
		return errors.New("no data")
	}
	if _, err := e.Reply().Textf(ctx, "Events in last 30 minutes: %s\nTotal: %s",
		formatInt(int(count.Row(0))),
		formatInt(int(total.Row(0))),
	); err != nil {
		return errors.Wrap(err, "reply")
	}

	return nil
}
