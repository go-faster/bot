package gh

import (
	"net/http"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type StatusWebhook struct {
	Meta            StatusMeta      `json:"meta"`
	Page            StatusPage      `json:"page"`
	ComponentUpdate ComponentUpdate `json:"component_update"`
	Component       StatusComponent `json:"component"`
}

type StatusMeta struct {
	Unsubscribe   string `json:"unsubscribe"`
	Documentation string `json:"documentation"`
}

type StatusPage struct {
	ID                string `json:"id"`
	StatusIndicator   string `json:"status_indicator"`
	StatusDescription string `json:"status_description"`
}

type ComponentUpdate struct {
	CreatedAt   time.Time `json:"created_at"`
	NewStatus   string    `json:"new_status"`
	OldStatus   string    `json:"old_status"`
	ID          string    `json:"id"`
	ComponentID string    `json:"component_id"`
}

type StatusComponent struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
}

func formatStatus(s StatusWebhook) message.StyledTextOption {
	formatter := func(eb *entity.Builder) error {
		eb.Plain("GitHub ")
		eb.Bold(s.Component.Name)
		eb.Plain(" status changed to ")
		eb.Bold(s.Component.Status)
		return nil
	}

	return styling.Custom(formatter)
}

func (w *Webhook) handleStatus(c echo.Context) error {
	// Handle Atlassian status webhook.
	//
	// See https://support.atlassian.com/statuspage/docs/enable-webhook-notifications/
	// GitHub status page: https://www.githubstatus.com

	ctx := c.Request().Context()
	ctx, span := w.tracer.Start(ctx, "github.status")
	defer span.End()

	var s StatusWebhook
	if err := c.Bind(&s); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return errors.Wrap(err, "bind")
	}

	if s.Component.Name == "" {
		// Incident update, ignoring.
		zctx.From(ctx).Debug("Ignoring incident update")
		return c.String(http.StatusOK, "ok")
	}

	span.AddEvent("StatusWebhook",
		trace.WithAttributes(
			attribute.String("component_name", s.Component.Name),
			attribute.String("component_status", s.Component.Status),
		),
	)

	// Notify telegram group.
	p, err := w.notifyPeer(ctx)
	if err != nil {
		return errors.Wrap(err, "peer")
	}
	if _, err := w.sender.To(p).
		NoWebpage().
		StyledText(ctx, formatStatus(s)); err != nil {
		return errors.Wrap(err, "send")
	}

	return c.String(http.StatusOK, "ok")
}
