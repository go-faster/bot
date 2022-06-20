package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func runMetrics(ctx context.Context, registry *prometheus.Registry, logger *zap.Logger) error {
	metricsAddr := os.Getenv("METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = "localhost:8081"
	}
	mux := http.NewServeMux()
	attachProfiler(mux)
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	server := &http.Server{Addr: metricsAddr, Handler: mux}

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		logger.Info("ListenAndServe", zap.String("addr", server.Addr))
		return server.ListenAndServe()
	})
	grp.Go(func() error {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		logger.Info("Shutdown", zap.String("addr", server.Addr))
		if err := server.Shutdown(shutCtx); err != nil {
			return multierr.Append(err, server.Close())
		}
		return nil
	})

	return grp.Wait()
}
