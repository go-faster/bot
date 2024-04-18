package api

import (
	"context"

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

func (s Server) NewError(ctx context.Context, err error) *oas.ErrorStatusCode {
	return &oas.ErrorStatusCode{
		StatusCode: 500,
		Response: oas.Error{
			Message: err.Error(),
		},
	}
}

var _ oas.Handler = (*Server)(nil)
