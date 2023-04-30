package gh

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v52/github"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/ent/check"
	"github.com/go-faster/bot/internal/ent/repository"
)

// Webhook is a Github events web hook handler.
type Webhook struct {
	db *ent.Client

	sender      *message.Sender
	notifyGroup string

	ghSecret string
	ghClient *github.Client
	ghID     int64

	updater *updater

	events metric.Int64Counter
	tracer trace.Tracer
	cache  *redis.Client
}

// NewWebhook creates new web hook handler.
func NewWebhook(
	db *ent.Client,
	gh *github.Client,
	ghID int64,
	sender *message.Sender,
	meterProvider metric.MeterProvider,
	tracerProvider trace.TracerProvider,
) *Webhook {
	meter := meterProvider.Meter("github.com/go-faster/bot/internal/gh/webhook")
	eventCount, err := meter.Int64Counter("github_event_count",
		metric.WithDescription("GitHub event counts"),
	)
	if err != nil {
		panic(err)
	}
	w := &Webhook{
		db:       db,
		sender:   sender,
		ghClient: gh,
		ghID:     ghID,
		events:   eventCount,
		tracer:   tracerProvider.Tracer("github.com/go-faster/bot/internal/gh/webhook"),
	}
	w.updater = newUpdater(w, 5*time.Second)
	return w
}

// Run runs some background tasks of Webhook.
func (h *Webhook) Run(ctx context.Context) error {
	if err := h.updater.Run(ctx); err != nil {
		return errors.Wrap(err, "PR updater")
	}
	return nil
}

func (h *Webhook) HasSecret() bool {
	return h.ghSecret != ""
}

func (h *Webhook) WithSecret(v string) *Webhook {
	h.ghSecret = v
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

// RegisterRoutes registers hook using given Echo router.
func (h *Webhook) RegisterRoutes(e *echo.Echo) {
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

func (h *Webhook) Handle(ctx context.Context, t string, data []byte) (rerr error) {
	// Normalize event type to match X-Github-Event value.
	if v, ok := _eventTypeToWebhookType[t]; ok {
		t = v
	}

	now := time.Now()

	ctx, span := h.tracer.Start(ctx, "wh.Handle",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	attrs := []attribute.KeyValue{
		attribute.String("event", t),
	}
	meta, err := extractEventMeta(data)
	if err != nil {
		zctx.From(ctx).Error("Failed to extract event meta",
			zap.String("type", t),
			zap.Error(err),
		)
	}
	attrs = append(attrs, meta.Attributes()...)
	span.SetAttributes(attrs...)
	defer func() {
		if rerr != nil {
			attrs = append(attrs, attribute.String("status", "error"))
		} else {
			attrs = append(attrs, attribute.String("status", "ok"))
		}
		h.events.Add(ctx, 1, metric.WithAttributes(attrs...))
	}()

	fields := []zap.Field{
		zap.String("type", t),
	}
	fields = append(fields, meta.Fields()...)
	ctx = zctx.With(ctx, fields...)
	lg := zctx.From(ctx)

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
			lg.Info("Unknown event type",
				zap.String("type", t),
			)
			span.SetStatus(codes.Ok, "ignored")
			return nil
		}
		return errors.Wrap(err, "parse")
	}

	lg.Info("Processing event")
	span.SetAttributes(
		attribute.String("event.go.type", fmt.Sprintf("%T", event)),
	)
	if err := h.processEvent(ctx, event); err != nil {
		return errors.Wrap(err, "process")
	}

	{
		// Update latest event timestamp.
		tx, err := h.db.BeginTx(ctx, nil)
		if err != nil {
			return errors.Wrap(err, "begin tx")
		}
		defer func() {
			_ = tx.Rollback()
		}()
		if err := tx.Organization.Create().
			SetID(meta.OrganizationID).
			SetName(meta.Organization).
			OnConflict(
				sql.ConflictColumns(check.FieldID),
				sql.ResolveWithNewValues(),
			).DoNothing().Exec(ctx); err != nil {
			return errors.Wrap(err, "upsert organization")
		}
		if err := tx.Repository.Create().
			SetID(meta.RepositoryID).
			SetName(meta.Repository).
			SetFullName(path.Join(meta.Organization, meta.Repository)).
			OnConflict(
				sql.ConflictColumns(check.FieldID),
				sql.ResolveWithNewValues(),
			).DoNothing().Exec(ctx); err != nil {
			return errors.Wrap(err, "upsert repository")
		}
		if err := tx.Repository.Update().Where(
			repository.ID(meta.RepositoryID),
			repository.Or(
				repository.LastEventAtIsNil(),
				repository.LastEventAtLT(now),
			),
		).SetLastEventAt(now).Exec(ctx); err != nil {
			return errors.Wrap(err, "update repository")
		}
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "commit")
		}
	}

	return nil
}

func (h *Webhook) handleHook(e echo.Context) error {
	ctx, span := h.tracer.Start(e.Request().Context(), "wh.handleHook",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	r := e.Request().WithContext(ctx)
	defer span.End()

	id := e.Request().Header.Get("X-GitHub-Delivery")
	if id == "" {
		zctx.From(ctx).Debug("No delivery ID")
		span.SetStatus(codes.Error, "no delivery id")
		return echo.ErrNotFound
	}
	span.SetAttributes(attribute.String("delivery_id", id))
	cacheKey := fmt.Sprintf("gh:delivery:%s", id)

	if h.cache != nil {
		// Check if we already processed this event.
		// Don't fail entire request if cache is failing.
		exists, err := h.cache.Exists(ctx, cacheKey).Result()
		if err != nil {
			zctx.From(ctx).Error("Failed to check cache",
				zap.Error(err),
			)
		}
		if exists == 1 {
			zctx.From(ctx).Debug("Already processed",
				zap.String("id", id),
			)
			span.SetStatus(codes.Ok, "hit")
			return e.String(http.StatusOK, "hit")
		}
	}

	payload, err := github.ValidatePayload(r, []byte(h.ghSecret))
	if err != nil {
		zctx.From(ctx).Debug("Failed to validate payload")
		span.SetStatus(codes.Error, err.Error())
		return echo.ErrNotFound
	}
	if err := h.Handle(ctx, github.WebHookType(r), payload); err != nil {
		zctx.From(ctx).Error("Failed to handle",
			zap.Error(err),
		)
		span.SetStatus(codes.Error, err.Error())
		return echo.ErrInternalServerError
	}

	if h.cache != nil {
		if err := h.cache.Set(ctx, cacheKey, 1, time.Hour).Err(); err != nil {
			zctx.From(ctx).Error("Failed to set cache",
				zap.Error(err),
			)
		}
	}

	span.SetStatus(codes.Ok, "done")
	return e.String(http.StatusOK, "done")
}

func (h *Webhook) processEvent(ctx context.Context, event interface{}) (rerr error) {
	lg := zctx.From(ctx)

	evType := fmt.Sprintf("%T", event)
	evType = strings.TrimPrefix(evType, "*github.")
	ctx, span := h.tracer.Start(ctx, fmt.Sprintf("wh.processEvent: %s", evType),
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attribute.String("e", evType)),
	)
	defer span.End()

	defer func() {
		if rerr != nil {
			span.RecordError(rerr)
			span.SetStatus(codes.Error, rerr.Error())
		} else {
			span.SetStatus(codes.Ok, "Done")
		}
	}()

	switch e := event.(type) {
	case *github.PullRequestEvent:
		return h.handlePR(ctx, e)
	case *github.ReleaseEvent:
		return h.handleRelease(ctx, e)
	case *github.RepositoryEvent:
		return h.handleRepo(ctx, e)
	case *github.IssuesEvent:
		return h.handleIssue(ctx, e)
	case *github.DiscussionEvent:
		return h.handleDiscussion(ctx, e)
	case *github.StarEvent:
		return h.handleStar(ctx, e)
	case *github.CheckRunEvent:
		return h.handleCheckRun(ctx, e)
	case *github.CheckSuiteEvent:
		return h.handleCheckSuite(ctx, e)
	case *github.WorkflowRunEvent:
		return h.handleWorkflowRun(ctx, e)
	case *github.WorkflowJobEvent:
		return h.handleWorkflowJob(ctx, e)
	default:
		lg.Info("No handler")
		return nil
	}
}

func (h *Webhook) notifyPeer(ctx context.Context) (tg.InputPeerClass, error) {
	p, err := h.sender.ResolveDomain(h.notifyGroup, peer.OnlyChannel).AsInputPeer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "resolve")
	}
	return p, nil
}

func (h *Webhook) WithCache(c *redis.Client) *Webhook {
	h.cache = c
	return h
}
