package gh

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/google/go-github/v42/github"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
)

func getDiscussionType(d *github.Discussion) string {
	cat := d.GetDiscussionCategory()
	emoji := cat.GetEmoji()
	if emoji == "" {
		return cat.GetName()
	}
	return cat.GetName() + " " + cat.GetEmoji()
}

func formatDiscussion(e *github.DiscussionEvent) message.StyledTextOption {
	discussion := e.GetDiscussion()
	sender := e.GetSender()
	formatter := func(eb *entity.Builder) error {
		eb.Plain("New ")
		eb.Plain(getDiscussionType(discussion))
		eb.Plain(" discussion")

		urlName := fmt.Sprintf(" %s#%d",
			e.GetRepo().GetFullName(),
			discussion.GetNumber(),
		)
		eb.TextURL(urlName, discussion.GetHTMLURL())
		eb.Plain(" by ")
		eb.TextURL(sender.GetLogin(), sender.GetHTMLURL())
		eb.Plain("\n\n")

		eb.Italic(discussion.GetTitle())
		eb.Plain("\n\n")

		return nil
	}

	return styling.Custom(formatter)
}

func (h Webhook) handleDiscussion(ctx context.Context, e *github.DiscussionEvent) error {
	if e.GetAction() != "created" {
		return nil
	}

	p, err := h.notifyPeer(ctx)
	if err != nil {
		return errors.Wrap(err, "peer")
	}

	if _, err := h.sender.To(p).NoWebpage().StyledText(ctx, formatDiscussion(e)); err != nil {
		return errors.Wrap(err, "send")
	}
	return nil
}
