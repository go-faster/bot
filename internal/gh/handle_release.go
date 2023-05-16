package gh

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/google/go-github/v52/github"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/td/telegram/message/styling"
)

func (w *Webhook) handleRelease(ctx context.Context, e *github.ReleaseEvent) error {
	ctx, span := w.tracer.Start(ctx, "handleRelease",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	if e.GetAction() != "published" {
		return nil
	}

	p, err := w.notifyPeer(ctx)
	if err != nil {
		return errors.Wrap(err, "peer")
	}

	if _, err := w.sender.To(p).StyledText(ctx,
		styling.Plain("New release: "),
		styling.TextURL(e.GetRelease().GetTagName(), e.GetRelease().GetHTMLURL()),
		styling.Plain(fmt.Sprintf(" for %s", e.GetRepo().GetFullName())),
	); err != nil {
		return errors.Wrap(err, "send")
	}

	return nil
}
