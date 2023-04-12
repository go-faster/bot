package gh

import (
	"context"

	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h Webhook) handleWorkflowRun(ctx context.Context, e *github.WorkflowRun) error {
	ctx, span := h.tracer.Start(ctx, "handleWorkflowRun",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("WorkflowRun",
		trace.WithStackTrace(true),
		trace.WithAttributes(
			attribute.String("name", e.GetName()),
			attribute.String("status", e.GetStatus()),
			attribute.String("conclusion", e.GetConclusion()),
			attribute.String("head_sha", e.GetHeadSHA()),
			attribute.String("event", e.GetEvent()),

			attribute.String("repository.full_name", e.GetRepository().GetFullName()),
		),
	)

	return nil
}
