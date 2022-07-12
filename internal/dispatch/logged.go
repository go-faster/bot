package dispatch

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// LoggedDispatcher is update logging middleware.
type LoggedDispatcher struct {
	handler telegram.UpdateHandler
	log     *zap.Logger
	tracer  trace.Tracer
}

// NewLoggedDispatcher creates new update logging middleware.
func NewLoggedDispatcher(next telegram.UpdateHandler, log *zap.Logger, traceProvider trace.TracerProvider) LoggedDispatcher {
	return LoggedDispatcher{
		handler: next,
		log:     log,
		tracer:  traceProvider.Tracer("td.dispatch.logged"),
	}
}

// Handle implements telegram.UpdateHandler.
func (d LoggedDispatcher) Handle(ctx context.Context, u tg.UpdatesClass) error {
	d.log.Debug("Update",
		zap.String("t", fmt.Sprintf("%T", u)),
	)
	ctx, span := d.tracer.Start(ctx, "handle: "+u.TypeName(),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(),
	)
	defer span.End()
	return d.handler.Handle(ctx, u)
}
