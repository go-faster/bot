package gh

import (
	"context"

	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (h *Webhook) handleCheckSuite(ctx context.Context, e *github.CheckSuiteEvent) error {
	_, span := h.tracer.Start(ctx, "handleCheckSuite",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	suite := e.GetCheckSuite()
	span.AddEvent("CheckSuiteEvent",
		trace.WithStackTrace(true),
		trace.WithAttributes(
			attribute.String("action", e.GetAction()),
			attribute.String("check_suite.status", suite.GetStatus()),
			attribute.String("check_suite.conclusion", suite.GetConclusion()),
			attribute.String("check_suite.head_sha", suite.GetHeadSHA()),

			attribute.Int64("organization.id", e.GetOrg().GetID()),
			attribute.String("organization.login", e.GetOrg().GetLogin()),
			attribute.String("repository.full_name", e.GetRepo().GetFullName()),
			attribute.Int64("repository.id", e.GetRepo().GetID()),
		),
	)

	ctx = zctx.With(ctx,
		zap.String("action", e.GetAction()),
		zap.Int64("check_suite.id", suite.GetID()),
		zap.String("head_sha", suite.GetHeadSHA()),
	)
	lg := zctx.From(ctx)

	var pr *github.PullRequest
	for _, pr = range suite.PullRequests {
		break
	}
	if pr == nil {
		// No PR - no update.
		lg.Debug("Ignore event: no PR info")
		return nil
	}
	lg.Debug("Emit check_update",
		zap.Int("pr.number", pr.GetNumber()),
		zap.String("pr.head_sha", pr.GetHead().GetSHA()),
	)

	h.updater.Emit(PullRequestUpdate{
		Event:  "check_update",
		Action: "",
		Repo:   e.GetRepo(),
		PR:     pr,
		Checks: nil,
	})
	return nil
}
