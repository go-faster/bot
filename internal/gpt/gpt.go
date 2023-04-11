package gpt

import (
	"context"
	"time"

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
func New(api *openai.Client, client *tg.Client) Handler {
	return Handler{api: api}
}

// OnMessage implements dispatch.MessageHandler.
func (h Handler) OnMessage(ctx context.Context, e dispatch.MessageEvent) error {
	return e.WithReply(ctx, func(reply *tg.Message) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		defer func() { _ = e.TypingAction().Cancel(ctx) }()
		sendTyping := func() { _ = e.TypingAction().Typing(ctx) }
		sendTyping()
		go func() {
			ticker := time.NewTicker(time.Second * 2)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					sendTyping()
				case <-ctx.Done():
					return
				}
			}
		}()

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
