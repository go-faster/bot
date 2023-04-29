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

func (h *Webhook) handleRepo(ctx context.Context, e *github.RepositoryEvent) error {
	ctx, span := h.tracer.Start(ctx, "handleRepo",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	switch e.GetAction() {
	case "created", "publicized":
		p, err := h.notifyPeer(ctx)
		if err != nil {
			return errors.Wrap(err, "peer")
		}

		if _, err := h.sender.To(p).StyledText(ctx,
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
