package gpt

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/go-faster/errors"
	"github.com/go-faster/simon/sdk/zctx"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/tg"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/ent/gptdialog"
)

// LimitConfig sets limits for GPT.
type LimitConfig struct {
	// PerUserRate sets cooldown timeout for a one user across chats.
	// If <=0, limit is disabled.
	PerUserRate time.Duration
	// PerPeerRate sets cooldown timeout for a one peer (chat/channel/user).
	// If <=0, limit is disabled.
	PerPeerRate time.Duration
	// MessageSizeLimit sets limit in runes for one message (prompt).
	// If <=0, limit is disabled.
	MessageSizeLimit int
	// DialogDepthLimit sets dialog depth limit. Counts prompt and AI answer as well.
	// If <=0, limit is disabled.
	DialogDepthLimit int
}

func parseEnv[T any](name string, target *T, parser func(string) (T, error)) error {
	val, ok := os.LookupEnv(name)
	if !ok {
		return nil
	}

	d, err := parser(val)
	if err != nil {
		return errors.Wrapf(err, "parse %q", name)
	}

	*target = d
	return nil
}

// ParseEnv parses environment.
func (cfg *LimitConfig) ParseEnv() error {
	if err := parseEnv("GPT_PER_USER_RATE", &cfg.PerUserRate, time.ParseDuration); err != nil {
		return err
	}

	if err := parseEnv("GPT_PER_PEER_RATE", &cfg.PerPeerRate, time.ParseDuration); err != nil {
		return err
	}

	if err := parseEnv("GPT_MESSAGE_SIZE_LIMIT", &cfg.MessageSizeLimit, strconv.Atoi); err != nil {
		return err
	}

	if err := parseEnv("GPT_DIALOG_DEPTH_LIMIT", &cfg.DialogDepthLimit, strconv.Atoi); err != nil {
		return err
	}

	return nil
}

func (cfg LimitConfig) setupLimiters(r *rateLimiters) {
	if prate := cfg.PerPeerRate; prate > 0 {
		r.peerLimiter = newLimiterMap(func(string) *rate.Limiter {
			return rate.NewLimiter(rate.Every(prate), 1)
		})
	}
	if urate := cfg.PerPeerRate; urate > 0 {
		r.userLimiter = newLimiterMap(func(int64) *rate.Limiter {
			return rate.NewLimiter(rate.Every(urate), 1)
		})
	}
}

type rateLimiters struct {
	peerLimiter *limiterMap[string]
	userLimiter *limiterMap[int64]
}

// Handler implements GPT request handler.
type Handler struct {
	db     *ent.Client
	api    *openai.Client
	tracer trace.Tracer

	contextPrompt *template.Template

	rateLimiters
	limitCfg LimitConfig
}

// New creates new Handler.
func New(api *openai.Client, db *ent.Client, tp trace.TracerProvider) *Handler {
	return &Handler{
		api:    api,
		db:     db,
		tracer: tp.Tracer("gpt"),
	}
}

// WithContextPromptTemplate sets template to setup a context prompt before completion.
func (h *Handler) WithContextPromptTemplate(t *template.Template) *Handler {
	h.contextPrompt = t
	return h
}

// WithMessageLimit sets message limit in runes.
func (h *Handler) WithLimitConfig(cfg LimitConfig) *Handler {
	h.limitCfg = cfg
	h.limitCfg.setupLimiters(&h.rateLimiters)
	return h
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

	reply := e.Message
	replyHdr, ok := reply.GetReplyTo()
	if !ok {
		return nil
	}

	lg := zctx.From(ctx).With(
		zap.String("reply", reply.Message),
		zap.Int("reply_to_msg_id", replyHdr.ReplyToMsgID),
		zap.Int("top_thread_id", replyHdr.ReplyToTopID),
	)
	ctx = zctx.With(ctx, lg)

	tx, err := h.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	switch exist, err := tx.GPTDialog.Query().
		Where(gptdialog.GptMsgID(replyHdr.ReplyToMsgID)).
		Exist(ctx); {
	case err != nil:
		return err
	case !exist:
		lg.Info("Do not answer to reply to message which is generated not by bot")
		return nil
	}

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

	dialogLimit := h.limitCfg.DialogDepthLimit
	switch l := len(dialog); {
	case l < 1:
		// No dialog, exit.
		return nil
	case dialogLimit > 0 && l > dialogLimit:
		if _, err := e.Reply().Text(ctx, "Dialog depth limit exceeded. Create a new thread by using gpt command."); err != nil {
			return errors.Wrap(err, "send dialog limit error")
		}
		return nil
	}

	if err := h.generateCompletion(ctx, e, reply, tx.GPTDialog, dialog, topMsgID); err != nil {
		return errors.Wrap(err, "generate completion")
	}

	return tx.Commit()
}

// OnMessage implements dispatch.MessageHandler.
func (h *Handler) OnCommand(ctx context.Context, e dispatch.MessageEvent) error {
	return e.WithReply(ctx, func(reply *tg.Message) error {
		return h.generateCompletion(ctx, e, reply, h.db.GPTDialog, nil, nil)
	})
}

func createPeerID(p tg.PeerClass) (peerID string, _ bool) {
	switch p := p.(type) {
	case *tg.PeerChannel:
		peerID = fmt.Sprintf("channel_%d", p.ChannelID)
	case *tg.PeerChat:
		peerID = fmt.Sprintf("chat_%d", p.ChatID)
	case *tg.PeerUser:
		peerID = fmt.Sprintf("user_%d", p.UserID)
	default:
		return peerID, false
	}
	return peerID, true
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

	if fromUser, ok := e.MessageFrom(); ok {
		if delay, ok := h.userLimiter.Allow(fromUser.ID); !ok {
			if _, err := e.Reply().Textf(ctx, "Per-user rate limit exceeded. Try again later (%s).", delay); err != nil {
				return errors.Wrap(err, "send user rate limit error")
			}
			return nil
		}
	}

	peerID, ok := createPeerID(reply.PeerID)
	if !ok {
		return errors.Errorf("unexpected input peer type %T", reply.PeerID)
	}

	if delay, ok := h.peerLimiter.Allow(peerID); !ok {
		if _, err := e.Reply().Textf(ctx, "Per-peer rate limit exceeded. Try again later (%s).", delay); err != nil {
			return errors.Wrap(err, "send peer rate limit error")
		}
		return nil
	}

	prompt := reply.GetMessage()
	if msgLimit := h.limitCfg.MessageSizeLimit; msgLimit > 0 &&
		utf8.RuneCountInString(prompt) > msgLimit {
		if _, err := e.Reply().Text(ctx, "Message is too big."); err != nil {
			return errors.Wrap(err, "send message limit error")
		}
		return nil
	}

	if t := h.contextPrompt; t != nil {
		data := generateContextPromptData(e)

		var sb strings.Builder
		if err := t.Execute(&sb, data); err != nil {
			zctx.From(ctx).Error("Context prompt execution error", zap.Error(err))
		} else {
			dialog = append([]openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: sb.String(),
				},
			}, dialog...)
		}
	}

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
