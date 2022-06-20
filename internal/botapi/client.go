package botapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"
)

// Client is simplified Telegram BotAPI client.
type Client struct {
	httpClient *http.Client
	token      string
}

// NewClient creates new Client.
func NewClient(token string, opts Options) *Client {
	opts.setDefaults()
	return &Client{
		httpClient: opts.HTTPClient,
		token:      token,
	}
}

func (m *Client) sendBotAPI(ctx context.Context, u string, result interface{}) (rErr error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return errors.Wrap(err, "create request")
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "send request")
	}
	defer multierr.AppendInvoke(&rErr, multierr.Close(resp.Body))

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return errors.Wrap(err, "decode json")
	}

	return nil
}

// GetFile sends getFile request to BotAPI.
func (m *Client) GetFile(ctx context.Context, id string) (rErr error) {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", m.token, url.QueryEscape(id))
	var result struct {
		OK          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code"`
		Description string `json:"description"`
	}
	if err := m.sendBotAPI(ctx, u, &result); err != nil {
		return errors.Wrap(err, "send")
	}
	if !result.OK {
		return errors.Errorf("API error %d: %s", result.ErrorCode, result.Description)
	}

	return nil
}

type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

// GetBotAPIMessage sends getUpdates request to BotAPI and finds message by msg_id.
//
// NB: it can find only recently received messages.
func (m *Client) GetBotAPIMessage(ctx context.Context, msgID int) (Message, error) {
	u := fmt.Sprintf(`https://api.telegram.org/bot%s/getUpdates?allowed_updates="message"`, m.token)
	var resp struct {
		OK          bool     `json:"ok"`
		ErrorCode   int      `json:"error_code"`
		Description string   `json:"description"`
		Result      []Update `json:"result"`
	}
	if err := m.sendBotAPI(ctx, u, &resp); err != nil {
		return Message{}, errors.Wrap(err, "send")
	}
	if !resp.OK {
		return Message{}, errors.Errorf("API error %d: %s", resp.ErrorCode, resp.Description)
	}

	for _, update := range resp.Result {
		if update.Message.MessageID == msgID {
			return update.Message, nil
		}
	}

	return Message{}, errors.Errorf("message %d not found", msgID)
}

// GetFileIDFromMessage finds file_id in message attachments.
func GetFileIDFromMessage(msg Message) (string, bool) {
	if len(msg.Photo) > 0 {
		return msg.Photo[0].FileID, true
	}
	switch {
	case msg.Animation.FileID != "":
		return msg.Animation.FileID, true
	case msg.Audio.FileID != "":
		return msg.Audio.FileID, true
	case msg.Document.FileID != "":
		return msg.Document.FileID, true
	case msg.Sticker.FileID != "":
		return msg.Sticker.FileID, true
	case msg.Video.FileID != "":
		return msg.Video.FileID, true
	case msg.VideoNote.FileID != "":
		return msg.VideoNote.FileID, true
	case msg.Voice.FileID != "":
		return msg.Voice.FileID, true
	}

	return "", false
}
