package api

import (
	"context"

	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/oas"
)

func NewServer(db *ent.Client) *Server {
	return &Server{
		db: db,
	}
}

type Server struct {
	db *ent.Client
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
