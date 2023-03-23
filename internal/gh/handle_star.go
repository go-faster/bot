package gh

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/google/go-github/v50/github"
	"github.com/gotd/td/telegram/message/styling"
	"go.uber.org/zap"
)

func (h Webhook) handleStar(ctx context.Context, e *github.StarEvent) error {
	if a := e.GetAction(); a != "created" {
		h.logger.Debug("Skipping action", zap.String("action", a))
		return nil
	}
	p, err := h.notifyPeer(ctx)
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
	if _, err := h.sender.To(p).NoWebpage().StyledText(ctx, options...); err != nil {
		return errors.Wrap(err, "send")
	}
	return nil
}
