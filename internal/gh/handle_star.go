package gh

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v52/github"
	"github.com/gotd/td/telegram/message/styling"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (w *Webhook) handleStar(ctx context.Context, e *github.StarEvent) error {
	ctx, span := w.tracer.Start(ctx, "handleStar",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	if a := e.GetAction(); a != "created" {
		zctx.From(ctx).Debug("Skipping action", zap.String("action", a))
		return nil
	}
	p, err := w.notifyPeer(ctx)
	if err != nil {
		return errors.Wrap(err, "peer")
	}

	var options []styling.StyledTextOption
	repo := e.GetRepo()
	sender := e.GetSender()
	options = append(options,
		styling.Plain("⭐ "),
		styling.TextURL(repo.GetFullName(), repo.GetHTMLURL()),
		styling.Bold(fmt.Sprintf(" %d ", repo.GetStargazersCount())),
		styling.Plain("by "),
	)
	options = append(options, styling.TextURL(sender.GetLogin(), sender.GetHTMLURL()))
	if name := sender.GetName(); name != "" {
		options = append(options, styling.Plain(fmt.Sprintf(" (%s)", name)))
	}
	if _, err := w.sender.To(p).NoWebpage().StyledText(ctx, options...); err != nil {
		return errors.Wrap(err, "send")
	}
	return nil
}
