package gpt

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
	"github.com/sashabaranov/go-openai"

	"github.com/go-faster/bot/internal/dispatch"
)

// Handler implements GPT request handler.
type Handler struct {
	api *openai.Client
}

// New creates new Handler.
func New(api *openai.Client) Handler {
	return Handler{api: api}
}

// OnMessage implements dispatch.MessageHandler.
func (h Handler) OnMessage(ctx context.Context, e dispatch.MessageEvent) error {
	return e.WithReply(ctx, func(reply *tg.Message) error {
		prompt := reply.GetMessage()
		resp, err := h.api.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model: openai.GPT3Dot5Turbo,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: prompt,
					},
				},
			},
		)
		if err != nil {
			if _, err := e.Reply().Text(ctx, "GPT server request failed"); err != nil {
				return errors.Wrap(err, "send")
			}
			return errors.Wrap(err, "send GPT request")
		}
		_, err = e.Reply().StyledText(ctx,
			styling.Bold(prompt),
			styling.Plain("\n"),
			styling.Plain(resp.Choices[0].Message.Content),
		)
		return err
	})
}
