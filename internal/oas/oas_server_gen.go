// Code generated by ogen, DO NOT EDIT.

package oas

import (
	"context"
)

// Handler handles operations described by OpenAPI v3 specification.
type Handler interface {
	// GetTelegramBadge implements getTelegramBadge operation.
	//
	// Get svg badge for telegram group.
	//
	// GET /badge/telegram/{group_name}
	GetTelegramBadge(ctx context.Context, params GetTelegramBadgeParams) (*SVGHeaders, error)
	// GetTelegramOnlineBadge implements getTelegramOnlineBadge operation.
	//
	// GET /badge/telegram/online
	GetTelegramOnlineBadge(ctx context.Context, params GetTelegramOnlineBadgeParams) (*SVGHeaders, error)
	// GithubStatus implements githubStatus operation.
	//
	// Https://www.githubstatus.com/ webhook.
	//
	// POST /github/status
	GithubStatus(ctx context.Context, req StatusNotification, params GithubStatusParams) error
	// Status implements status operation.
	//
	// Get status.
	//
	// GET /status
	Status(ctx context.Context) (*Status, error)
	// NewError creates *ErrorStatusCode from error returned by handler.
	//
	// Used for common default response.
	NewError(ctx context.Context, err error) *ErrorStatusCode
}

// Server implements http server based on OpenAPI v3 specification and
// calls Handler to handle requests.
type Server struct {
	h Handler
	baseServer
}

// NewServer creates new Server.
func NewServer(h Handler, opts ...ServerOption) (*Server, error) {
	s, err := newServerConfig(opts...).baseServer()
	if err != nil {
		return nil, err
	}
	return &Server{
		h:          h,
		baseServer: s,
	}, nil
}
