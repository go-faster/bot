package gh

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/go-faster/simon/sdk/zctx"
	"github.com/google/go-github/v50/github"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/styling"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-faster/bot/internal/action"
)

type issueType string

const (
	featureRequest issueType = "feature request"
	bugReport      issueType = "bug report"
	plain          issueType = "issue"
)

func getIssueType(issue *github.Issue) issueType {
	for _, label := range issue.Labels {
		switch label.GetName() {
		case "enhancement":
			return featureRequest
		case "bug":
			return bugReport
		}
	}

	return plain
}

func formatIssue(e *github.IssuesEvent) message.StyledTextOption {
	issue := e.GetIssue()
	user := issue.GetUser()
	formatter := func(eb *entity.Builder) error {
		eb.Plain("New ")
		eb.Plain(string(getIssueType(issue)))

		urlName := fmt.Sprintf(" %s#%d",
			e.GetRepo().GetFullName(),
			issue.GetNumber(),
		)
		eb.TextURL(urlName, issue.GetHTMLURL())
		eb.Plain(" by ")
		eb.TextURL(user.GetLogin(), user.GetHTMLURL())
		eb.Plain("\n\n")

		eb.Italic(issue.GetTitle())
		eb.Plain("\n\n")

		length := len(issue.Labels)
		if length > 0 {
			eb.Italic("Labels: ")

			for idx, label := range issue.Labels {
				switch label.GetName() {
				case "":
					continue
				case "bug":
					eb.Bold(label.GetName())
				default:
					eb.Italic(label.GetName())
				}

				if idx != length-1 {
					eb.Plain(", ")
				}
			}
		}
		return nil
	}

	return styling.Custom(formatter)
}

func (h *Webhook) handleIssue(ctx context.Context, e *github.IssuesEvent) error {
	ctx, span := h.tracer.Start(ctx, "handleIssue",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	if e.GetAction() != "opened" {
		zctx.From(ctx).Info("Ignoring non-opened issue")
		return nil
	}

	p, err := h.notifyPeer(ctx)
	if err != nil {
		return errors.Wrap(err, "peer")
	}

	if _, err := h.sender.To(p).NoWebpage().
		Row(
			markup.Callback("Close",
				action.Marshal(action.Action{
					Type:   "close",
					ID:     e.GetIssue().GetNumber(),
					Entity: "issue",
				}),
			),
		).
		StyledText(ctx, formatIssue(e)); err != nil {
		return errors.Wrap(err, "send")
	}
	return nil
}
