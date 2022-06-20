package storage

import (
	"strconv"

	"github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v42/github"
	"go.uber.org/multierr"
)

// PRMsgIDKey generates key for given PR.
func PRMsgIDKey(pr *github.PullRequestEvent) []byte {
	key := strconv.AppendInt([]byte("pr_"), pr.GetRepo().GetID(), 10)
	key = strconv.AppendInt(key, int64(pr.GetPullRequest().GetNumber()), 10)
	return key
}

// LastMsgIDKey generates last message ID key for given channel.
func LastMsgIDKey(channelID int64) []byte {
	return strconv.AppendInt([]byte("last_msg_"), channelID, 10)
}

// MsgID is a simple message ID storage.
type MsgID struct {
	db *pebble.DB
}

// NewMsgID creates new MsgID.
func NewMsgID(db *pebble.DB) MsgID {
	return MsgID{db: db}
}

// UpdateLastMsgID updates last message ID for given channel.
func (m MsgID) UpdateLastMsgID(channelID int64, msgID int) (rerr error) {
	key := LastMsgIDKey(channelID)

	b := m.db.NewIndexedBatch()
	data, closer, err := b.Get(key)
	switch {
	case errors.Is(err, pebble.ErrNotFound):
	case err != nil:
		return err
	default:
		defer func() {
			multierr.AppendInto(&rerr, closer.Close())
		}()
		s := string(data)
		id, err := strconv.Atoi(s)
		if err != nil {
			return errors.Wrapf(err, "parse msg id %q", s)
		}

		if id > msgID {
			return nil
		}
	}

	if err := b.Set(key, strconv.AppendInt(nil, int64(msgID), 10), pebble.Sync); err != nil {
		return errors.Wrapf(err, "set msg_id %d", channelID)
	}

	if err := b.Commit(nil); err != nil {
		return errors.Wrap(err, "commit")
	}

	return nil
}

// SetPRNotification sets PR notification message ID.
func (m MsgID) SetPRNotification(pr *github.PullRequestEvent, msgID int) error {
	return m.db.Set(PRMsgIDKey(pr), strconv.AppendInt(nil, int64(msgID), 10), pebble.Sync)
}

// FindPRNotification finds PR notification message ID and last message ID for given channel.
// NB: even if last message ID was not found, function returns non-zero msgID.
func (m MsgID) FindPRNotification(channelID int64, pr *github.PullRequestEvent) (msgID, lastMsgID int, rerr error) {
	prID := pr.GetPullRequest().GetNumber()
	snap := m.db.NewSnapshot()
	defer func() {
		multierr.AppendInto(&rerr, snap.Close())
	}()

	var err error
	msgID, err = findInt(snap, PRMsgIDKey(pr))
	if err != nil {
		return 0, 0, errors.Wrapf(err, "find msg ID of PR #%d notification", prID)
	}

	lastMsgID, err = findInt(snap, LastMsgIDKey(channelID))
	if err != nil {
		return msgID, 0, errors.Wrapf(err, "find last msg ID of channel %d", channelID)
	}

	return msgID, lastMsgID, nil
}

func findInt(snap *pebble.Snapshot, key []byte) (_ int, rerr error) {
	data, closer, err := snap.Get(key)
	if err != nil {
		return 0, err
	}
	defer func() {
		multierr.AppendInto(&rerr, closer.Close())
	}()

	s := string(data)
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.Wrapf(err, "parse msg id %q", s)
	}

	return id, nil
}
