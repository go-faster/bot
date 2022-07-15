package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Entry struct {
	User string
	Date time.Time
}

func reset(columns ...proto.Column) {
	for _, c := range columns {
		c.Reset()
	}
}

func run(ctx context.Context, lg *zap.Logger) error {
	var (
		filePath     = flag.String("f", "", "path to file")
		decodeBuffer = flag.Int("b", 1024, "json decode buffer size")
	)
	flag.Parse()
	f, err := os.Open(*filePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	g, ctx := errgroup.WithContext(ctx)

	// Usernames mapping for consistency.
	names := make(map[string]string)

	entries := make(chan Entry, 1)
	g.Go(func() error {
		j := jx.Decode(f, *decodeBuffer)
		defer close(entries)

		return j.Obj(func(d *jx.Decoder, key string) error {
			switch key {
			case "messages":
				return d.Arr(func(d *jx.Decoder) error {
					var (
						date     time.Time
						userID   string
						userName string
					)
					if err := d.Obj(func(d *jx.Decoder, key string) error {
						switch key {
						case "date":
							s, err := d.Str()
							if err != nil {
								return errors.Wrap(err, "date")
							}
							if date, err = time.Parse("2006-01-02T15:04:05", s); err != nil {
								lg.Warn("failed to parse date",
									zap.String("date", s),
									zap.Error(err),
								)
								return nil
							}
							return nil
						case "from_id":
							if userID, err = d.Str(); err != nil {
								return errors.Wrap(err, "from_id")
							}
							return nil
						case "from":
							if userName, err = d.Str(); err != nil {
								return errors.Wrap(err, "from")
							}
							return nil
						default:
							return d.Skip()
						}
					}); err != nil {
						return errors.Wrap(err, "messages")
					}
					if userID == "" || date.IsZero() {
						return nil
					}
					if name, ok := names[userID]; ok {
						userName = name
					} else {
						names[userID] = userName
					}
					select {
					case entries <- Entry{User: userName, Date: date}:
						return nil
					case <-ctx.Done():
						return ctx.Err()
					}
				})
			default:
				return d.Skip()
			}
		})
	})
	g.Go(func() error {
		colDate := new(proto.ColDate)
		colUser := new(proto.ColStr).LowCardinality()
		db, err := ch.Dial(ctx, ch.Options{})
		if err != nil {
			return errors.Wrap(err, "dial")
		}
		/*
			CREATE DATABASE tg;
			CREATE TABLE tg.messages (
			    `date` Date,
			    `user` LowCardinality(String)
			) engine = MergeTree ORDER BY (`user`, `date`);
		*/
		if err := db.Do(ctx, ch.Query{
			Body: "INSERT INTO tg.messages (date, user) VALUES",
			Input: proto.Input{
				{Name: "date", Data: colDate},
				{Name: "user", Data: colUser},
			},
			OnInput: func(ctx context.Context) error {
				reset(colDate, colUser)
				for {
					if colDate.Rows() > 10000 {
						return nil
					}
					select {
					case <-ctx.Done():
						return ctx.Err()
					case e, ok := <-entries:
						if !ok {
							return nil
						}
						colDate.Append(e.Date)
						colUser.Append(e.User)
					}
				}
			},
		}); err != nil {
			return errors.Wrap(err, "query")
		}
		return nil
	})
	return g.Wait()
}

func main() {
	ctx := context.Background()
	lg, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	if err := run(ctx, lg); err != nil {
		lg.Fatal("failed to run", zap.Error(err))
	}
}
