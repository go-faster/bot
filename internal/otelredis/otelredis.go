package otelredis

import (
	"context"
	"net"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

type Hook struct {
	tracer trace.Tracer
}

func NewHook(tp trace.TracerProvider) *Hook {
	return &Hook{
		tracer: tp.Tracer("redis"),
	}
}

func (h Hook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h Hook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		ctx, span := h.tracer.Start(ctx, "redis: "+cmd.Name(),
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()
		return next(ctx, cmd)
	}
}

func (h Hook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}

var _ redis.Hook = (*Hook)(nil)
