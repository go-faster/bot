package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/povilasv/prommod"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/go-faster/bot/internal/metrics"
)

func run(ctx context.Context) error {
	logger, _ := zap.NewProduction(
		zap.IncreaseLevel(zapcore.DebugLevel),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	defer func() { _ = logger.Sync() }()

	registry := prometheus.NewPedanticRegistry()
	mts := metrics.NewMetrics()
	registry.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
		prommod.NewCollector("gotdbot"),
		mts,
	)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return runBot(ctx, mts, logger.Named("bot"))
	})
	g.Go(func() error {
		return runMetrics(ctx, registry, logger.Named("metrics"))
	})

	return g.Wait()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
