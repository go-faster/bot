package gh

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v42/github"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/tg"
)

func getPullRequestURL(e *github.PullRequestEvent) styling.StyledTextOption {
	urlName := fmt.Sprintf("%s#%d",
		e.GetRepo().GetFullName(),
		e.PullRequest.GetNumber(),
	)

	return styling.TextURL(urlName, e.GetPullRequest().GetHTMLURL())
}

func getPullRequestAuthor(e *github.PullRequestEvent) styling.StyledTextOption {
	u := e.GetPullRequest().GetUser()
	return styling.TextURL(u.GetLogin(), u.GetHTMLURL())
}

func getPullRequestMergedBy(e *github.PullRequestEvent) styling.StyledTextOption {
	u := e.GetPullRequest().GetMergedBy()
	return styling.TextURL(u.GetLogin(), u.GetHTMLURL())
}

func (h Webhook) notifyPR(p tg.InputPeerClass, e *github.PullRequestEvent) *message.Builder {
	r := h.sender.To(p).NoWebpage()
	if u, _ := url.Parse(e.GetPullRequest().GetHTMLURL()); u != nil {
		files, checks := *u, *u
		files.Path = path.Join(files.Path, "files")
		checks.Path = path.Join(checks.Path, "checks")
		r = r.Row(
			markup.URL("Diff🔀", files.String()),
			markup.URL("Checks▶", checks.String()),
		)
	}
	return r
}

func (h Webhook) handlePRClosed(ctx context.Context, e *github.PullRequestEvent) error {
	prID := e.GetPullRequest().GetNumber()
	log := h.logger.With(zap.Int("pr", prID), zap.String("repo", e.GetRepo().GetFullName()))
	if !e.GetPullRequest().GetMerged() {
		h.logger.Info("Ignoring non-merged PR")
		return nil
	}

	p, err := h.notifyPeer(ctx)
	if err != nil {
		return errors.Wrap(err, "peer")
	}

	var replyID int
	fallback := func(ctx context.Context) error {
		r := h.notifyPR(p, e)
		if replyID != 0 {
			r = r.Reply(replyID)
		}
		if _, err := r.StyledText(ctx,
			styling.Plain("Pull request "),
			getPullRequestURL(e),
			styling.Plain(" merged by "),
			getPullRequestMergedBy(e),
			styling.Plain("\n\n"),
			styling.Italic(e.GetPullRequest().GetTitle()),
		); err != nil {
			return errors.Wrap(err, "send")
		}

		return nil
	}

	ch, ok := p.(*tg.InputPeerChannel)
	if !ok {
		return fallback(ctx)
	}

	msgID, lastMsgID, err := h.storage.FindPRNotification(ch.ChannelID, e)
	if msgID != 0 {
		log.Debug("Found PR notification ID", zap.Int("msg_id", msgID))
		replyID = msgID
	}
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return fallback(ctx)
		}
		return errors.Wrap(err, "find notification")
	}

	log.Debug("Found last message ID", zap.Int("msg_id", lastMsgID), zap.Int64("channel", ch.ChannelID))
	if lastMsgID-msgID > 10 {
		log.Debug("Can't merge, send new message")
		return fallback(ctx)
	}

	if _, err := h.notifyPR(p, e).Edit(msgID).StyledText(ctx,
		styling.Plain("Pull request "),
		getPullRequestURL(e),
		styling.Plain(" "),
		styling.Strike("opened by "),
		styling.Custom(func(eb *entity.Builder) error {
			u := e.GetPullRequest().GetUser()
			eb.Format(
				u.GetLogin(),
				entity.Strike(),
				entity.TextURL(u.GetHTMLURL()),
			)
			return nil
		}),
		styling.Plain(" merged by "),
		getPullRequestMergedBy(e),
		styling.Plain("\n\n"),
		styling.Italic(e.GetPullRequest().GetTitle()),
	); err != nil {
		return errors.Wrap(err, "send")
	}

	return nil
}

func (h Webhook) handlePROpened(ctx context.Context, event *github.PullRequestEvent) error {
	p, err := h.notifyPeer(ctx)
	if err != nil {
		return errors.Wrap(err, "peer")
	}
	action := " opened"
	if event.GetPullRequest().GetDraft() {
		action = " drafted"
	}

	msgID, err := unpack.MessageID(h.notifyPR(p, event).StyledText(ctx,
		styling.Plain("New pull request "),
		getPullRequestURL(event),
		styling.Plain(action),
		styling.Plain(" by "),
		getPullRequestAuthor(event),
		styling.Plain("\n\n"),
		styling.Italic(event.GetPullRequest().GetTitle()),
	))
	if err != nil {
		return errors.Wrap(err, "send")
	}

	ch, ok := p.(*tg.InputPeerChannel)
	if !ok {
		return h.storage.SetPRNotification(event, msgID)
	}

	return multierr.Append(
		h.storage.UpdateLastMsgID(ch.ChannelID, msgID),
		h.storage.SetPRNotification(event, msgID),
	)
}

func (h Webhook) handlePR(ctx context.Context, e *github.PullRequestEvent) error {
	// Ignore PR-s from dependabot (too much noise).
	// TODO(ernado): delay and merge into single message
	if e.GetPullRequest().GetUser().GetLogin() == "dependabot[bot]" {
		h.logger.Info("Ignored PR from dependabot")
		return nil
	}

	switch e.GetAction() {
	case "opened":
		return h.handlePROpened(ctx, e)
	case "closed":
		return h.handlePRClosed(ctx, e)
	}
	return nil
}
