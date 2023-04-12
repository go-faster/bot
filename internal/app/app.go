// Package app implements OTEL, prometheus, graceful shutdown and other common application features
// for pfm projects.
package app

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/go-faster/simon/sdk/zctx"
)

const (
	exitCodeOk             = 0
	exitCodeApplicationErr = 1
	exitCodeWatchdog       = 1
)

const (
	shutdownTimeout = time.Second * 5
	watchdogTimeout = shutdownTimeout + time.Second*5
)

// Run f until interrupt.
//
// If errors.Is(err, ctx.Err()) is valid for returned error, shutdown is considered graceful.
// Context is cancelled on SIGINT. After watchdogTimeout application is forcefully terminated
// with exitCodeWatchdog.
func Run(f func(ctx context.Context, lg *zap.Logger, m *Metrics) error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := zap.NewProductionConfig()
	if s := os.Getenv("OTEL_LOG_LEVEL"); s != "" {
		var lvl zapcore.Level
		if err := lvl.UnmarshalText([]byte(s)); err != nil {
			panic(err)
		}
		cfg.Level.SetLevel(lvl)
	}
	lg, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	// Add logger to root context.
	ctx = zctx.With(ctx, lg)

	m, err := newMetrics(ctx, lg.Named("metrics"))
	if err != nil {
		panic(err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer lg.Info("Shutting down")
		if err := f(ctx, lg, m); err != nil {
			if errors.Is(err, ctx.Err()) {
				// Parent context got cancelled, error is expected.
				lg.Debug("Graceful shutdown")
				return nil
			}
			return err
		}

		// Also shutting down metrics server to stop error group.
		cancel()

		return nil
	})
	g.Go(func() error {
		if err := m.run(ctx); err != nil {
			// Should already handle context cancellation gracefully.
			return errors.Wrap(err, "metrics")
		}
		return nil
	})

	go func() {
		// Guaranteed way to kill application.
		// Helps if f is stuck, e.g. deadlock during shutdown.
		<-ctx.Done()

		// Context is canceled, giving application time to shut down gracefully.

		lg.Info("Waiting for application shutdown")
		time.Sleep(watchdogTimeout)

		// Application is not shutting down gracefully, kill it.
		// This code should not be executed if f is already returned.

		lg.Warn("Graceful shutdown watchdog triggered: forcing shutdown")
		os.Exit(exitCodeWatchdog)
	}()

	if err := g.Wait(); err != nil {
		lg.Error("Failed", zap.Error(err))
		os.Exit(exitCodeApplicationErr)
	}

	lg.Info("Application stopped")
	os.Exit(exitCodeOk)
}
