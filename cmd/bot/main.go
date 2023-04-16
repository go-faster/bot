package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/app"
)

func main() {
	println("start")
	defer func() {
		if re := recover(); re != nil {
			_, _ = fmt.Fprintf(os.Stderr, "panic: %v", re)
		}
		time.Sleep(time.Second * 3)
	}()
	app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) error {
		return runBot(ctx, m, lg.Named("bot"))
	})
}
