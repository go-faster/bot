package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	"github.com/dustin/go-humanize"
	"github.com/go-faster/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/dispatch"
)

func formatInt(x int) string {
	return formatFloat(float64(x))
}

func formatFloat(num float64) string {
	v, u := humanize.ComputeSI(num)
	return fmt.Sprintf("%.1f%s", v, u)
}

func (a *App) clickHouse(ctx context.Context) (db *ch.Client, cleanup func(), err error) {
	db, err = ch.Dial(ctx, ch.Options{
		Address:        os.Getenv("CLICKHOUSE_ADDR"),
		Compression:    ch.CompressionZSTD,
		TracerProvider: a.m.TracerProvider(),
		MeterProvider:  a.m.MeterProvider(),
		Database:       "faster",

		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		User:     os.Getenv("CLICKHOUSE_USER"),

		OpenTelemetryInstrumentation: true,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "connect")
	}
	cleanup = func() {
		if err := db.Close(); err != nil {
			a.lg.Error("Close clickhouse", zap.Error(err))
		}
	}
	return db, cleanup, nil
}

func (a *App) HandleEvents(ctx context.Context, e dispatch.MessageEvent) error {
	db, cleanup, err := a.clickHouse(ctx)
	if err != nil {
		return errors.Wrap(err, "clickHouse")
	}
	defer cleanup()
	var count, total proto.ColUInt64
	if err := db.Do(ctx, ch.Query{
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
	if err := db.Do(ctx, ch.Query{
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

func (a *App) FetchEvents(ctx context.Context, start time.Time) error {
	r := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	db, cleanup, err := a.clickHouse(ctx)
	if err != nil {
		return errors.Wrap(err, "clickHouse")
	}
	defer cleanup()

	q := fmt.Sprintf("SELECT (id, ts, event) FROM faster.github_events_raw WHERE ts >= toDateTime64(%d, 9) ORDER BY ", start.Unix())
	var (
		colID   proto.ColUInt64
		colTime proto.ColDateTime64
		colBody proto.ColStr
	)
	if err := db.Do(ctx, ch.Query{
		Body: q,
		Result: proto.Results{
			{Name: "id", Data: &colID},
			{Name: "ts", Data: &colTime},
			{Name: "event", Data: &colBody},
		},
		OnResult: func(ctx context.Context, block proto.Block) error {
			for i := 0; i < colID.Rows(); i++ {
				var (
					id = colID.Row(i)
					ts = colTime.Row(i)
					b  = colBody.RowBytes(i)
				)

				h := sha256.Sum256(b)
				if _, err := r.Set(ctx, fmt.Sprintf("event:%d:%d", ts.Unix(), id), h[:], time.Minute).Result(); err != nil {
					return errors.Wrap(err, "set")
				}
			}
			return nil
		},
	}); err != nil {
		return errors.Wrap(err, "do")
	}

	return nil
}
