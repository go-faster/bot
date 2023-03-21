// Package zctx is a context-aware zap logger.
package zctx

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type key struct{}

var _nop = zap.NewNop()

// From returns zap.Logger from context.
func From(ctx context.Context) *zap.Logger {
	v, ok := ctx.Value(key{}).(*zap.Logger)
	if !ok || v == nil {
		return _nop
	}
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		// Hex-encoded lowercase as per OTEL log data model.
		v = v.With(
			zap.Stringer("trace_id", spanCtx.TraceID()),
			zap.Stringer("span_id", spanCtx.SpanID()),
		)
	}
	return v
}

// With returns new context with provided zap.Logger.
func With(ctx context.Context, lg *zap.Logger) context.Context {
	return context.WithValue(ctx, key{}, lg)
}
