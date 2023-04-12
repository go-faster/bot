package gh

import (
	"context"

	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h Webhook) handleCheckSuite(ctx context.Context, e *github.CheckSuiteEvent) error {
	ctx, span := h.tracer.Start(ctx, "handleCheckSuite",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("CheckSuiteEvent",
		trace.WithStackTrace(true),
		trace.WithAttributes(
			attribute.String("action", e.GetAction()),
			attribute.String("check_suite.status", e.GetCheckSuite().GetStatus()),
			attribute.String("check_suite.conclusion", e.GetCheckSuite().GetConclusion()),
			attribute.String("check_suite.head_sha", e.GetCheckSuite().GetHeadSHA()),

			attribute.Int64("organization.id", e.GetOrg().GetID()),
			attribute.String("organization.login", e.GetOrg().GetLogin()),
			attribute.String("repository.full_name", e.GetRepo().GetFullName()),
			attribute.Int64("repository.id", e.GetRepo().GetID()),
		),
	)

	return nil
}
