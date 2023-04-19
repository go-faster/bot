package gh

import (
	"context"

	"github.com/go-faster/sdk/zctx"
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

	lg := zctx.From(ctx)
	var pr *github.PullRequest
	for _, pr = range e.GetWorkflowRun().PullRequests {
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
