package api

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/ogen-go/ogen/http"

	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/oas"
)

func NewServer(db *ent.Client, tg *telegram.Client, resolver peer.Resolver, ht http.Client) *Server {
	return &Server{
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
