package app

import (
	"context"

	"github.com/google/go-github/v50/github"
)

type Storage interface {
	UpdateLastMsgID(ctx context.Context, channelID int64, msgID int) (rerr error)
	SetPRNotification(ctx context.Context, pr *github.PullRequestEvent, msgID int) error
	FindPRNotification(ctx context.Context, channelID int64, pr *github.PullRequestEvent) (msgID, lastMsgID int, rerr error)
}
