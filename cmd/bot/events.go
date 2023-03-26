package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	"github.com/dustin/go-humanize"
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/gh"
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
	begin := time.Now()
	ctx, span := a.tracer.Start(ctx, "FetchEvents",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	trackedRepo := map[string]struct{}{
		"ClickHouse/ch-go":      {},
		"ernado/oss-estimator":  {},
		"ernado/du":             {},
		"VKCOM/statshouse":      {},
		"anonyindian/gotgproto": {},
		"xelaj/mtproto":         {},
		"xelaj/tl":              {},
	}

	r := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	db, cleanup, err := a.clickHouse(ctx)
	if err != nil {
		return errors.Wrap(err, "clickHouse")
	}
	defer cleanup()

	q := fmt.Sprintf("SELECT id, ts, raw FROM faster.github_events_raw WHERE ts >= toDateTime64(%d, 9) ORDER BY ts DESC, id DESC", start.Unix())
	var (
		colID   proto.ColInt64
		colTime proto.ColDateTime
		colBody proto.ColStr
		d       jx.Decoder
		e       jx.Encoder

		total     int
		skipped   int
		hit       int
		processed int
		sizeTotal uint64
		latest    time.Time
	)
	if err := db.Do(ctx, ch.Query{
		Body: q,
		Result: proto.Results{
			{Name: "id", Data: &colID},
			{Name: "ts", Data: &colTime},
			{Name: "raw", Data: &colBody},
		},
		OnResult: func(ctx context.Context, block proto.Block) error {
			for i := 0; i < colID.Rows(); i++ {
				var (
					id = colID.Row(i)
					b  = colBody.RowBytes(i)
					t  = colTime.Row(i)
				)
				if t.After(latest) {
					latest = t
				}
				sizeTotal += uint64(len(b))
				d.ResetBytes(b)
				e.Reset()
				ev, err := gh.Transform(&d, &e)
				if err != nil {
					return errors.Wrap(err, "transform")
				}
				d.ResetBytes(e.Bytes())
				if err := d.Validate(); err != nil {
					return errors.Wrap(err, "validate")
				}
				total++
				if _, ok := trackedRepo[ev.RepoName]; !ok {
					skipped++
					continue
				}

				// Protect from duplicates.
				k := fmt.Sprintf("v3:event:%x", id)
				exists, err := r.Exists(ctx, k).Result()
				if err != nil {
					return errors.Wrap(err, "exists")
				}
				if exists != 0 {
					hit++
					continue
				}
				if _, err := r.Set(ctx, k, 1, time.Hour).Result(); err != nil {
					return errors.Wrap(err, "set")
				}
				a.lg.Info("Got event",
					zap.Int64("id", id),
					zap.String("type", ev.Type),
					zap.Int64("repo_id", ev.RepoID),
					zap.String("repo_name", ev.RepoName),
				)

				// Handle as event was received from webhook.
				if err := a.wh.Handle(ctx, ev.Type, e.Bytes()); err != nil {
					return errors.Wrap(err, "handle")
				}

				processed++
			}
			return nil
		},
	}); err != nil {
		return errors.Wrap(err, "do")
	}

	a.lg.Info("FetchEvents",
		zap.String("duration.human", time.Since(begin).String()),
		zap.String("lag.human", time.Since(latest).String()),
		zap.Int("total", total),
		zap.Int("skipped", skipped),
		zap.Int("hit", hit),
		zap.Int("processed", processed),
		zap.String("parsed", humanize.Bytes(sizeTotal)),
	)

	return nil
}
