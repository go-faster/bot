package gh

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/google/go-github/v42/github"
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

	repo := e.GetRepo()
	if _, err := h.sender.To(p).NoWebpage().StyledText(ctx,
		styling.Plain("New star: "),
		styling.TextURL(repo.GetFullName(), repo.GetHTMLURL()),
		styling.Plain(fmt.Sprintf(" %d ⭐", repo.GetStargazersCount())),
	); err != nil {
		return errors.Wrap(err, "send")
	}
	return nil
}
