package state

import (
	"context"

	"github.com/google/go-github/v50/github"
)

type Check struct {
	ID         int64
	Name       string // e.g. "build"
	Conclusion string // failure, success, neutral, cancelled, skipped, timed_out, action_required
	Status     string // completed
}

type Storage interface {
	UpdateLastMsgID(ctx context.Context, channelID int64, msgID int) error
	SetPRNotification(ctx context.Context, pr *github.PullRequestEvent, msgID int) error
	FindPRNotification(ctx context.Context, channelID int64, pr *github.PullRequestEvent) (msgID, lastMsgID int, rerr error)
	UpsertCheck(ctx context.Context, check *github.CheckRunEvent) ([]Check, error)
}
