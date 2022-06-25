package app

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/td/fileid"
	"github.com/gotd/td/tg"

	"github.com/go-faster/bot/internal/botapi"
)

// checkOurFileID generates file_id and tries to use it in BotAPI.
func (m Middleware) checkOurFileID(ctx context.Context, id fileid.FileID) error {
	if m.client == nil {
		return nil
	}

	encoded, err := fileid.EncodeFileID(id)
	if err != nil {
		return errors.Wrap(err, "encode")
	}

	if err := m.client.GetFile(ctx, encoded); err != nil {
		return errors.Wrap(err, "check file_id")
	}

	return nil
}

// tryGetFileID decodes file_id from BotAPI and tries to map into Telegram API file location.
func (m Middleware) tryGetFileID(ctx context.Context, msgID int) (tg.InputFileLocationClass, string, error) {
	botAPIMsg, err := m.client.GetBotAPIMessage(ctx, msgID)
	if err != nil {
		return nil, "", errors.Wrap(err, "get message")
	}

	encoded, ok := botapi.GetFileIDFromMessage(botAPIMsg)
	if !ok {
		return nil, "", errors.New("no media in message")
	}

	fileID, err := fileid.DecodeFileID(encoded)
	if err != nil {
		return nil, encoded, errors.Wrap(err, "decode file_id")
	}

	loc, ok := fileID.AsInputFileLocation()
	if !ok {
		return nil, encoded, errors.New("can't map to location")
	}

	return loc, encoded, nil
}
