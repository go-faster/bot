package gh

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-faster/errors"
	"github.com/google/go-github/v45/github"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/storage"
)

// Webhook is a Github events web hook handler.
type Webhook struct {
	storage storage.MsgID

	sender       *message.Sender
	notifyGroup  string
	githubSecret string

	logger *zap.Logger
	events syncint64.Counter
	tracer trace.Tracer
}

// NewWebhook creates new web hook handler.
func NewWebhook(
	msgID storage.MsgID,
	sender *message.Sender,
	githubSecret string,
	meterProvider metric.MeterProvider,
	tracerProvider trace.TracerProvider,
) *Webhook {
	meter := meterProvider.Meter("github.com/go-faster/bot/internal/gh/webhook")
	eventCount, err := meter.SyncInt64().Counter("github_event_count",
		instrument.WithDescription("GitHub event counts"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		panic(err)
	}
	return &Webhook{
		events:       eventCount,
		storage:      msgID,
		sender:       sender,
		githubSecret: githubSecret,
		logger:       zap.NewNop(),
		tracer:       tracerProvider.Tracer("github.com/go-faster/bot/internal/gh/webhook"),
	}
}

// WithSender sets message sender to use.
func (h *Webhook) WithSender(sender *message.Sender) *Webhook {
	h.sender = sender
	return h
}

// WithNotifyGroup sets channel name to send notifications.
func (h *Webhook) WithNotifyGroup(domain string) *Webhook {
	h.notifyGroup = domain
	return h
}

// WithLogger sets logger to use.
func (h *Webhook) WithLogger(logger *zap.Logger) *Webhook {
	h.logger = logger
	return h
}

// RegisterRoutes registers hook using given Echo router.
func (h Webhook) RegisterRoutes(e *echo.Echo) {
	e.POST("/hook", h.handleHook)
}

func (h Webhook) handleHook(e echo.Context) error {
	ctx, span := h.tracer.Start(e.Request().Context(), "handleHook",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	r := e.Request().WithContext(ctx)
	defer span.End()

	payload, err := github.ValidatePayload(r, []byte(h.githubSecret))
	if err != nil {
		h.logger.Info("Failed to validate payload")
		span.SetStatus(codes.Error, err.Error())
		return echo.ErrNotFound
	}
	whType := github.WebHookType(r)
	span.SetName("hook: " + whType)
	span.SetAttributes(attribute.String("github.webhook.type", whType))

	if whType == "security_advisory" {
		// Current GitHub library is unable to handle this.
		span.SetStatus(codes.Ok, "ignored")
		return e.String(http.StatusOK, "ignored")
	}

	event, err := github.ParseWebHook(whType, payload)
	if err != nil {
		h.logger.Error("Failed to parse webhook", zap.Error(err))
		span.SetStatus(codes.Error, err.Error())
		return echo.ErrInternalServerError
	}

	h.events.Add(ctx, 1,
		attribute.String("event", whType),
	)

	log := h.logger.With(
		zap.String("type", fmt.Sprintf("%T", event)),
	)
	log.Info("Processing event")
	if err := h.processEvent(ctx, event, log); err != nil {
		log.Error("Failed to process event", zap.Error(err))
		span.SetStatus(codes.Error, err.Error())
		return echo.ErrInternalServerError
	}
	span.SetStatus(codes.Ok, "done")
	return e.String(http.StatusOK, "done")
}

func (h Webhook) processEvent(ctx context.Context, event interface{}, log *zap.Logger) error {
	switch event := event.(type) {
	case *github.PullRequestEvent:
		return h.handlePR(ctx, event)
	case *github.ReleaseEvent:
		return h.handleRelease(ctx, event)
	case *github.RepositoryEvent:
		return h.handleRepo(ctx, event)
	case *github.IssuesEvent:
		return h.handleIssue(ctx, event)
	case *github.DiscussionEvent:
		return h.handleDiscussion(ctx, event)
	case *github.StarEvent:
		return h.handleStar(ctx, event)
	default:
		log.Info("No handler")
		return nil
	}
}

func (h Webhook) notifyPeer(ctx context.Context) (tg.InputPeerClass, error) {
	p, err := h.sender.ResolveDomain(h.notifyGroup, peer.OnlyChannel).AsInputPeer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "resolve")
	}
	return p, nil
}
