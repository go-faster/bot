package gh

import (
	"context"

	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v52/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (w *Webhook) handleWorkflowRun(ctx context.Context, e *github.WorkflowRunEvent) error {
	_, span := w.tracer.Start(ctx, "handleWorkflowRun",
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

	ctx = zctx.With(ctx,
		zap.String("action", e.GetAction()),
		zap.Int64("workflow_run.id", run.GetID()),
		zap.String("workflow_run.name", run.GetName()),
		zap.String("head_sha", run.GetHeadSHA()),
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
	lg.Debug("Emit check_update",
		zap.Int("pr.number", pr.GetNumber()),
		zap.String("pr.head_sha", pr.GetHead().GetSHA()),
	)

	return w.updater.Emit(PullRequestUpdate{
		Event:  "check_update",
		Action: "",
		Repo:   e.GetRepo(),
		PR:     pr,
		Checks: nil,
	})
}
