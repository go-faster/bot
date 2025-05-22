package api

import (
	"context"

	"github.com/go-faster/sdk/zctx"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/json"
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

func (s Server) GithubStatus(ctx context.Context, req oas.GithubStatusReq, params oas.GithubStatusParams) error {
	lg := zctx.From(ctx)
	lg.Info("Github status key", zap.String("key", params.Secret.Value))
	for k, v := range req {
		var object any
		_ = json.Unmarshal(v, &object)

		lg.Debug("github status",
			zap.String("key", k),
			zap.Any("value", object),
			zap.Stringer("valuer.raw", v),
		)
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
