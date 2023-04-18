package gh

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v50/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (h *Webhook) handleCheckRun(ctx context.Context, e *github.CheckRunEvent) error {
	_, span := h.tracer.Start(ctx, "handleCheckRun",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("CheckRunEvent",
		trace.WithStackTrace(true),
		trace.WithAttributes(
			attribute.String("action", e.GetAction()),
			attribute.String("check_run.name", e.GetCheckRun().GetName()),
			attribute.String("check_run.status", e.GetCheckRun().GetStatus()),
			attribute.String("check_run.conclusion", e.GetCheckRun().GetConclusion()),
			attribute.String("check_run.head_sha", e.GetCheckRun().GetHeadSHA()),

			attribute.Int64("organization.id", e.GetOrg().GetID()),
			attribute.String("organization.login", e.GetOrg().GetLogin()),
			attribute.String("repository.full_name", e.GetRepo().GetFullName()),
			attribute.Int64("repository.id", e.GetRepo().GetID()),
		),
	)

	ctx = zctx.With(ctx,
		zap.String("action", e.GetAction()),
		zap.Int64("run_id", e.GetCheckRun().GetID()),
		zap.String("head_sha", e.GetCheckRun().GetHeadSHA()),
	)
	lg := zctx.From(ctx)

	pr, err := h.upsertCheck(ctx, e)
	if err != nil {
		return errors.Wrap(err, "upsert check")
	}
	if pr == nil {
		// No PR - no update.
		lg.Debug("Ignore event: no PR info")
		return nil
	}
	lg.Debug("Update check",
		zap.String("pr_head_sha", pr.GetHead().GetSHA()),
	)

	// We gonna update all checks anyway, so do not include check id in key.
	key := fmt.Sprintf("%s#%d",
		e.GetRepo().GetFullName(),
		pr.GetNumber(),
	)
	_, err, _ = h.checksSema.Do(key, func() (any, error) {
		checks, err := h.queryChecks(ctx, e.GetRepo(), pr)
		if err != nil {
			return 0, errors.Wrap(err, "query checks")
		}

		return 0, h.updatePR(ctx, PullRequestUpdate{
			Event:  "check_run",
			Action: "",
			Repo:   e.GetRepo(),
			PR:     pr,
			Checks: checks,
		})
	})
	return err
}
