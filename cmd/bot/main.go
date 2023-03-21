package main

import (
	"context"

	"go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/app"
	"github.com/go-faster/bot/internal/otelenv"
)

func main() {
	otelenv.Set(
		semconv.ServiceNameKey.String("bot"),
		semconv.ServiceNamespaceKey.String("faster"),
	)
	app.Run(func(ctx context.Context, lg *zap.Logger, m *app.Metrics) error {
		return runBot(ctx, m, lg.Named("bot"))
	})
}
