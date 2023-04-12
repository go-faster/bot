package gh

import (
	"context"

	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h Webhook) handleWorkflowJob(ctx context.Context, e *github.WorkflowJob) error {
	ctx, span := h.tracer.Start(ctx, "handleWorkflowJob",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("WorkflowJob",
		trace.WithStackTrace(true),
		trace.WithAttributes(
			attribute.String("workflow.name", e.GetWorkflowName()),
			attribute.String("name", e.GetName()),
		),
	)

	return nil
}
