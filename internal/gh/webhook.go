package gh

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-faster/errors"
	"github.com/google/go-github/v50/github"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/state"
)

// Webhook is a Github events web hook handler.
type Webhook struct {
	storage state.Storage

	sender       *message.Sender
	notifyGroup  string
	githubSecret string

	logger *zap.Logger
	events instrument.Int64Counter
	tracer trace.Tracer
}

// NewWebhook creates new web hook handler.
func NewWebhook(
	msgID state.Storage,
	sender *message.Sender,
	meterProvider metric.MeterProvider,
	tracerProvider trace.TracerProvider,
) *Webhook {
	meter := meterProvider.Meter("github.com/go-faster/bot/internal/gh/webhook")
	eventCount, err := meter.Int64Counter("github_event_count",
		instrument.WithDescription("GitHub event counts"),
	)
	if err != nil {
		panic(err)
	}
	return &Webhook{
		events:  eventCount,
		storage: msgID,
		sender:  sender,
		logger:  zap.NewNop(),
		tracer:  tracerProvider.Tracer("github.com/go-faster/bot/internal/gh/webhook"),
	}
}

func (h *Webhook) HasSecret() bool {
	return h.githubSecret != ""
}

func (h *Webhook) WithSecret(v string) *Webhook {
	h.githubSecret = v
	return h
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

func eventMapping() map[string]string {
	return map[string]string{
		"branch_protection_rule":         "BranchProtectionRuleEvent",
		"check_run":                      "CheckRunEvent",
		"check_suite":                    "CheckSuiteEvent",
		"code_scanning_alert":            "CodeScanningAlertEvent",
		"commit_comment":                 "CommitCommentEvent",
		"content_reference":              "ContentReferenceEvent",
		"create":                         "CreateEvent",
		"delete":                         "DeleteEvent",
		"deploy_key":                     "DeployKeyEvent",
		"deployment":                     "DeploymentEvent",
		"deployment_status":              "DeploymentStatusEvent",
		"discussion":                     "DiscussionEvent",
		"fork":                           "ForkEvent",
		"github_app_authorization":       "GitHubAppAuthorizationEvent",
		"gollum":                         "GollumEvent",
		"installation":                   "InstallationEvent",
		"installation_repositories":      "InstallationRepositoriesEvent",
		"issue_comment":                  "IssueCommentEvent",
		"issues":                         "IssuesEvent",
		"label":                          "LabelEvent",
		"marketplace_purchase":           "MarketplacePurchaseEvent",
		"member":                         "MemberEvent",
		"membership":                     "MembershipEvent",
		"merge_group":                    "MergeGroupEvent",
		"meta":                           "MetaEvent",
		"milestone":                      "MilestoneEvent",
		"organization":                   "OrganizationEvent",
		"org_block":                      "OrgBlockEvent",
		"package":                        "PackageEvent",
		"page_build":                     "PageBuildEvent",
		"ping":                           "PingEvent",
		"project":                        "ProjectEvent",
		"project_card":                   "ProjectCardEvent",
		"project_column":                 "ProjectColumnEvent",
		"public":                         "PublicEvent",
		"pull_request":                   "PullRequestEvent",
		"pull_request_review":            "PullRequestReviewEvent",
		"pull_request_review_comment":    "PullRequestReviewCommentEvent",
		"pull_request_review_thread":     "PullRequestReviewThreadEvent",
		"pull_request_target":            "PullRequestTargetEvent",
		"push":                           "PushEvent",
		"repository":                     "RepositoryEvent",
		"repository_dispatch":            "RepositoryDispatchEvent",
		"repository_import":              "RepositoryImportEvent",
		"repository_vulnerability_alert": "RepositoryVulnerabilityAlertEvent",
		"release":                        "ReleaseEvent",
		"secret_scanning_alert":          "SecretScanningAlertEvent",
		"star":                           "StarEvent",
		"status":                         "StatusEvent",
		"team":                           "TeamEvent",
		"team_add":                       "TeamAddEvent",
		"user":                           "UserEvent",
		"watch":                          "WatchEvent",
		"workflow_dispatch":              "WorkflowDispatchEvent",
		"workflow_job":                   "WorkflowJobEvent",
		"workflow_run":                   "WorkflowRunEvent",
	}
}

func reverseMapping(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[v] = k
	}
	return out
}

var _eventTypeToWebhookType = reverseMapping(eventMapping())

func (h Webhook) Handle(ctx context.Context, t string, data []byte) (rerr error) {
	// Normalize event type to match X-Github-Event value.
	if v, ok := _eventTypeToWebhookType[t]; ok {
		t = v
	}

	ctx, span := h.tracer.Start(ctx, "wh.Handle",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	defer func() {
		if rerr != nil {
			span.SetStatus(codes.Error, rerr.Error())
		} else {
			span.SetStatus(codes.Ok, "Done")
		}
	}()

	if t == "security_advisory" {
		// Current GitHub library is unable to handle this.
		span.SetStatus(codes.Ok, "ignored")
		return nil
	}
	event, err := github.ParseWebHook(t, data)
	if err != nil {
		if strings.Contains(err.Error(), "unknown X-Github-Event") {
			h.logger.Info("Unknown event type",
				zap.String("type", t),
			)
			span.SetStatus(codes.Ok, "ignored")
			return nil
		}
		return errors.Wrap(err, "parse")
	}
	attr := attribute.String("event", t)
	span.SetAttributes(attr)
	h.events.Add(ctx, 1, attr)
	log := h.logger.With(
		zap.String("type", fmt.Sprintf("%T", event)),
	)
	log.Info("Processing event")
	if err := h.processEvent(ctx, event, log); err != nil {
		return errors.Wrap(err, "process")
	}

	return nil
}

func (h Webhook) handleHook(e echo.Context) error {
	ctx, span := h.tracer.Start(e.Request().Context(), "wh.handleHook",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	r := e.Request().WithContext(ctx)
	defer span.End()

	payload, err := github.ValidatePayload(r, []byte(h.githubSecret))
	if err != nil {
		h.logger.Debug("Failed to validate payload")
		span.SetStatus(codes.Error, err.Error())
		return echo.ErrNotFound
	}
	if err := h.Handle(ctx, github.WebHookType(r), payload); err != nil {
		h.logger.Error("Failed to handle",
			zap.Error(err),
		)
		span.SetStatus(codes.Error, err.Error())
		return echo.ErrInternalServerError
	}

	span.SetStatus(codes.Ok, "done")
	return e.String(http.StatusOK, "done")
}

func (h Webhook) processEvent(ctx context.Context, event interface{}, log *zap.Logger) error {
	evType := fmt.Sprintf("%T", event)
	evType = strings.TrimPrefix(evType, "*github.")
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("wh.processEvent: %s", evType),
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attribute.String("event", evType)),
	)
	defer span.End()

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
