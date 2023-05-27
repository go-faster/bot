package api

import (
	"context"

	"github.com/go-faster/bot/internal/oas"
)

type Server struct{}

func (s Server) Status(ctx context.Context) (*oas.Status, error) {
	return &oas.Status{
		Status:  oas.StatusStatusOk,
		Message: oas.NewOptString("All systems operational"),
	}, nil
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
