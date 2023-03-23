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
	"github.com/go-faster/jx"
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

	q := fmt.Sprintf("SELECT id, ts, raw FROM faster.github_events_raw WHERE ts >= toDateTime64(%d, 9) ORDER BY ts DESC, id DESC", start.Unix())
	var (
		colID   proto.ColUInt64
		colTime proto.ColDateTime64
		colBody proto.ColStr
		d       jx.Decoder
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
					ts = colTime.Row(i)
					b  = colBody.RowBytes(i)
				)
				d.ResetBytes(b)
				var (
					payload []byte
					t       string

					repoID   int64
					repoName string
				)
				if err := d.ObjBytes(func(d *jx.Decoder, key []byte) error {
					switch string(key) {
					case "payload":
						if payload, err = d.StrBytes(); err != nil {
							return errors.Wrap(err, "payload")
						}
						return nil
					case "type":
						if t, err = d.Str(); err != nil {
							return errors.Wrap(err, "type")
						}
						return nil
					case "repo":
						return d.ObjBytes(func(d *jx.Decoder, key []byte) error {
							switch string(key) {
							case "id":
								if repoID, err = d.Int64(); err != nil {
									return errors.Wrap(err, "id")
								}
								return nil
							case "name":
								if repoName, err = d.Str(); err != nil {
									return errors.Wrap(err, "name")
								}
								return nil
							default:
								return d.Skip()
							}
						})
					default:
						return d.Skip()
					}
				}); err != nil {
					return errors.Wrap(err, "decode")
				}
				d.ResetBytes(payload)
				if err := d.Validate(); err != nil {
					return errors.Wrap(err, "validate")
				}
				h := sha256.Sum256(payload)
				if _, err := r.Set(ctx, fmt.Sprintf("event:%s:%d:%d", t, ts.Unix(), id), h[:], time.Minute).Result(); err != nil {
					return errors.Wrap(err, "set")
				}
				a.lg.Debug("Event",
					zap.Int64("repo_id", repoID),
					zap.String("repo_name", repoName),
					zap.String("type", t),
					zap.Uint64("id", id),
					zap.String("sha256", fmt.Sprintf("%x", h[:])),
				)
			}
			return nil
		},
	}); err != nil {
		return errors.Wrap(err, "do")
	}

	return nil
}