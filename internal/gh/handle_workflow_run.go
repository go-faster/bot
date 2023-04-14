package gh

import (
	"context"

	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h *Webhook) handleWorkflowRun(ctx context.Context, e *github.WorkflowRunEvent) error {
	_, span := h.tracer.Start(ctx, "handleWorkflowRun",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	run := e.GetWorkflowRun()
	span.AddEvent("WorkflowRun",
		trace.WithStackTrace(true),
		trace.WithAttributes(
			attribute.String("name", run.GetName()),
			attribute.String("status", run.GetStatus()),
			attribute.String("conclusion", run.GetConclusion()),
			attribute.String("head_sha", run.GetHeadSHA()),
			attribute.String("event", run.GetEvent()),

			attribute.Int64("organization.id", e.GetOrg().GetID()),
			attribute.String("organization.login", e.GetOrg().GetLogin()),
			attribute.Int64("repository.id", e.GetRepo().GetID()),
			attribute.String("repository.full_name", e.GetRepo().GetFullName()),
		),
	)

	return nil
}
