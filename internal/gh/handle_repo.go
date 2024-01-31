package gh

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v52/github"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram/message/styling"
)

func (w *Webhook) handleRepo(ctx context.Context, e *github.RepositoryEvent) error {
	ctx, span := w.tracer.Start(ctx, "handleRepo",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	if e.GetRepo().GetPrivate() {
		zctx.From(ctx).Info("Private repository", zap.String("repo", e.GetRepo().GetFullName()))
		return nil
	}

	switch e.GetAction() {
	case "created", "publicized":
		p, err := w.notifyPeer(ctx)
		if err != nil {
			return errors.Wrap(err, "peer")
		}

		if _, err := w.sender.To(p).StyledText(ctx,
			styling.Plain("New repository "),
			styling.TextURL(e.GetRepo().GetFullName(), e.GetRepo().GetHTMLURL()),
		); err != nil {
			return errors.Wrap(err, "send")
		}

		return nil
	default:
		zctx.From(ctx).Info("Type ignored", zap.String("action", e.GetAction()))

		return nil
	}
}
