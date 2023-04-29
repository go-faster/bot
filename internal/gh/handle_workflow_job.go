package gh

import (
	"context"

	"github.com/google/go-github/v52/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h *Webhook) handleWorkflowJob(ctx context.Context, e *github.WorkflowJobEvent) error {
	_, span := h.tracer.Start(ctx, "handleWorkflowJob",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	j := e.GetWorkflowJob()

	span.AddEvent("WorkflowJob",
		trace.WithStackTrace(true),
		trace.WithAttributes(
			attribute.String("workflow.name", j.GetWorkflowName()),
			attribute.String("name", j.GetName()),
		),
	)

	return nil
}
