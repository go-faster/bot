package api

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/ogen-go/ogen/http"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/oas"
)

func NewServer(lg *zap.Logger, db *ent.Client, tg *telegram.Client, resolver peer.Resolver, ht http.Client) *Server {
	return &Server{
		lg:       lg,
		db:       db,
		tg:       tg,
		ht:       ht,
		resolver: resolver,
	}
}

type Server struct {
	db       *ent.Client
	tg       *telegram.Client
	ht       http.Client
	resolver peer.Resolver
	lg       *zap.Logger
}

func (s Server) GithubStatus(ctx context.Context, req oas.StatusNotification, params oas.GithubStatusParams) error {
	if params.Secret.Value == "" {
		return errors.New("not authenticated")
	}

	switch req.Type {
	case oas.StatusNotificationComponentUpdateStatusNotification:
		s.lg.Info("Github status: component update")
	case oas.StatusNotificationIncidentUpdateStatusNotification:
		s.lg.Info("Github status: incident update")
	default:
		return nil
	}

	return nil
}

func (s Server) NewError(ctx context.Context, err error) *oas.ErrorStatusCode {
	return &oas.ErrorStatusCode{
		StatusCode: 500,
		Response: oas.Error{
			Message: err.Error(),
		},
	}
}

var _ oas.Handler = (*Server)(nil)
