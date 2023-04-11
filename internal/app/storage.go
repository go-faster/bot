package app

import "github.com/google/go-github/v50/github"

type Storage interface {
	UpdateLastMsgID(channelID int64, msgID int) (rerr error)
	SetPRNotification(pr *github.PullRequestEvent, msgID int) error
	FindPRNotification(channelID int64, pr *github.PullRequestEvent) (msgID, lastMsgID int, rerr error)
}
