package dispatch

import (
	"context"

	"github.com/gotd/td/telegram/message/inline"
	"github.com/gotd/td/tg"
)

// InlineQuery represents inline query event.
type InlineQuery struct {
	QueryID  int64
	Query    string
	Offset   string
	Enquirer *tg.InputUser

	geo  *tg.GeoPoint
	user *tg.User

	baseEvent
}

// Reply returns result builder.
func (e InlineQuery) Reply() *inline.ResultBuilder {
	return inline.New(e.rpc, e.rand, e.QueryID)
}

// User returns User object if available.
// False and nil otherwise.
func (e InlineQuery) User() (*tg.User, bool) {
	return e.user, e.user != nil
}

// Geo returns GeoPoint object and true if query has attached geo point.
// False and nil otherwise.
func (e InlineQuery) Geo() (*tg.GeoPoint, bool) {
	return e.geo, e.geo != nil
}

// InlineHandler is a simple inline query event handler.
type InlineHandler interface {
	OnInline(ctx context.Context, e InlineQuery) error
}

// InlineHandlerFunc is a functional adapter for Handler.
type InlineHandlerFunc func(ctx context.Context, e InlineQuery) error

// OnInline implements InlineHandler.
func (h InlineHandlerFunc) OnInline(ctx context.Context, e InlineQuery) error {
	return h(ctx, e)
}
