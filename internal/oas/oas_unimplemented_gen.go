// Code generated by ogen, DO NOT EDIT.

package oas

import (
	"context"

	ht "github.com/ogen-go/ogen/http"
)

// UnimplementedHandler is no-op Handler which returns http.ErrNotImplemented.
type UnimplementedHandler struct{}

var _ Handler = UnimplementedHandler{}

// GetTelegramBadge implements getTelegramBadge operation.
//
// Get svg badge for telegram group.
//
// GET /badge/telegram/{group_name}
func (UnimplementedHandler) GetTelegramBadge(ctx context.Context, params GetTelegramBadgeParams) (r *SVGHeaders, _ error) {
	return r, ht.ErrNotImplemented
}

// GetTelegramOnlineBadge implements getTelegramOnlineBadge operation.
//
// GET /badge/telegram/online
func (UnimplementedHandler) GetTelegramOnlineBadge(ctx context.Context, params GetTelegramOnlineBadgeParams) (r *SVGHeaders, _ error) {
	return r, ht.ErrNotImplemented
}

// GithubStatus implements githubStatus operation.
//
// Https://www.githubstatus.com/ webhook.
//
// POST /github/status
func (UnimplementedHandler) GithubStatus(ctx context.Context, req StatusNotification, params GithubStatusParams) error {
	return ht.ErrNotImplemented
}

// Status implements status operation.
//
// Get status.
//
// GET /status
func (UnimplementedHandler) Status(ctx context.Context) (r *Status, _ error) {
	return r, ht.ErrNotImplemented
}

// NewError creates *ErrorStatusCode from error returned by handler.
//
// Used for common default response.
func (UnimplementedHandler) NewError(ctx context.Context, err error) (r *ErrorStatusCode) {
	r = new(ErrorStatusCode)
	return r
}
