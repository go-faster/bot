package gh

import (
	"context"

	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h *Webhook) handleCheckSuite(ctx context.Context, e *github.CheckSuiteEvent) error {
	_, span := h.tracer.Start(ctx, "handleCheckSuite",
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

	lg := zctx.From(ctx)
	var pr *github.PullRequest
	for _, pr = range e.GetCheckSuite().PullRequests {
		break
	}
	if pr == nil {
		// No PR - no update.
		lg.Debug("Ignore event: no PR info")
		return nil
	}

	h.updater.Emit(PullRequestUpdate{
		Event:  "check_update",
		Action: "",
		Repo:   e.GetRepo(),
		PR:     pr,
		Checks: nil,
	})
	return nil
}
