package gpt

import (
	"context"
	"fmt"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/simon/sdk/zctx"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/tg"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/ent/gptdialog"
)

// Handler implements GPT request handler.
type Handler struct {
	db     *ent.Client
	api    *openai.Client
	tracer trace.Tracer
}

// New creates new Handler.
func New(api *openai.Client, db *ent.Client, tp trace.TracerProvider) *Handler {
	return &Handler{api: api, db: db, tracer: tp.Tracer("gpt")}
}

// OnReply handles replies to gpt generated messages.
func (h *Handler) OnReply(ctx context.Context, e dispatch.MessageEvent) (rerr error) {
	ctx, span := h.tracer.Start(ctx, "OnReply")
	defer span.End()

	defer func() {
		if rerr != nil {
			span.RecordError(rerr)
			span.SetStatus(codes.Error, rerr.Error())
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
	}()

	lg := zctx.From(ctx)

	reply := e.Message

	replyHdr, ok := reply.GetReplyTo()
	if !ok {
		return nil
	}

	tx, err := h.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var (
		pred     = gptdialog.GptMsgID(replyHdr.ReplyToMsgID)
		topMsgID = &replyHdr.ReplyToMsgID
	)
	if threadTopID, ok := replyHdr.GetReplyToTopID(); ok {
		pred = gptdialog.Or(pred, gptdialog.ThreadTopMsgID(threadTopID))
		topMsgID = &threadTopID
	}

	thread, err := tx.GPTDialog.Query().
		Where(pred).
		Order(gptdialog.ByGptMsgID()).
		All(ctx)
	if err != nil {
		return err
	}

	var (
		firstMsg = zap.Skip()
		lastMsg  = zap.Skip()
	)
	if len(thread) > 0 {
		firstMsg = zap.String("first_msg", thread[0].PromptMsg)
		lastMsg = zap.String("last_msg", thread[len(thread)-1].GptMsg)
	}

	lg.Info("Query dialog",
		zap.Int("reply_to_msg_id", replyHdr.ReplyToMsgID),
		zap.Intp("top_msg_id", topMsgID),
		zap.Int("got", len(thread)),
		firstMsg,
		lastMsg,
	)

	var dialog []openai.ChatCompletionMessage
	for _, row := range thread {
		dialog = append(dialog,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: row.PromptMsg,
			},
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: row.GptMsg,
			},
		)
	}

	if len(dialog) > 0 {
		if err := h.generateCompletion(ctx, e, reply, tx.GPTDialog, dialog, topMsgID); err != nil {
			return errors.Wrap(err, "generate completion")
		}
	}

	return tx.Commit()
}

// OnMessage implements dispatch.MessageHandler.
func (h *Handler) OnCommand(ctx context.Context, e dispatch.MessageEvent) error {
	return e.WithReply(ctx, func(reply *tg.Message) error {
		return h.generateCompletion(ctx, e, reply, h.db.GPTDialog, nil, nil)
	})
}

func (h *Handler) generateCompletion(
	ctx context.Context,
	e dispatch.MessageEvent,
	reply *tg.Message,
	tx *ent.GPTDialogClient,
	dialog []openai.ChatCompletionMessage,
	topMsgId *int,
) error {
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
			Messages: append(dialog, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			}),
		},
	)
	if err != nil {
		if _, err := e.Reply().Text(ctx, "GPT server request failed"); err != nil {
			return errors.Wrap(err, "send error report")
		}
		return errors.Wrap(err, "send GPT request")
	}

	choices := resp.Choices
	if len(choices) < 1 {
		return errors.Wrap(err, "GPT returned no message")
	}
	gptMessage := choices[0].Message.Content

	msgID, err := unpack.MessageID(e.Reply().StyledText(ctx,
		styling.Bold(prompt),
		styling.Plain("\n"),
		styling.Plain(gptMessage),
	))
	if err != nil {
		return errors.Wrap(err, "send message")
	}

	var peerID string
	switch peer := reply.PeerID.(type) {
	case *tg.PeerChannel:
		peerID = fmt.Sprintf("channel_%d", peer.ChannelID)
	case *tg.PeerChat:
		peerID = fmt.Sprintf("chat_%d", peer.ChatID)
	case *tg.PeerUser:
		peerID = fmt.Sprintf("user_%d", peer.UserID)
	default:
		return errors.Errorf("unexpected input peer type %T", peer)
	}

	{
		// Save message to the dialog.
		b := tx.Create().
			SetPeerID(peerID).
			SetPromptMsg(prompt).
			SetPromptMsgID(reply.ID).
			SetGptMsg(gptMessage).
			SetGptMsgID(msgID).
			SetNillableThreadTopMsgID(topMsgId)
		if err := b.Exec(ctx); err != nil {
			return errors.Wrapf(err, "insert message %d", msgID)
		}
	}

	return nil
}
