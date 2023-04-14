package main

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/go-faster/simon/sdk/zctx"
	"github.com/gotd/td/tg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/action"
	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/ent/user"
)

func (a *App) OnButton(ctx context.Context, e dispatch.Button) error {
	ctx, span := a.tracer.Start(ctx, "OnBotCallbackQuery")
	defer span.End()
	if e.Input == nil {
		span.SetStatus(codes.Ok, "Ignored")
		zctx.From(ctx).Info("OnButton: no user")
		return nil
	}
	zctx.From(ctx).Info("OnButton",
		zap.String("user", e.Input.String()),
	)

	var act action.Action
	if err := act.UnmarshalText(e.Data); err != nil {
		answer := &tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: e.QueryID,
			Message: fmt.Sprintf("Error: %s", err),
			Alert:   true,
		}
		if _, err := e.RPC().MessagesSetBotCallbackAnswer(ctx, answer); err != nil {
			return err
		}
		span.SetStatus(codes.Error, err.Error())
		return nil
	}

	var hasToken bool
	{
		users, err := a.db.User.Query().Where(
			user.ID(e.User.ID),
		).All(ctx)
		if err != nil {
			return errors.Wrap(err, "query user")
		}
		for _, u := range users {
			if u.GithubToken != "" {
				hasToken = true
				break
			}
		}
	}

	span.SetAttributes(
		attribute.Int("action.id", act.ID),
		attribute.Stringer("action.entity", act.Entity),
		attribute.Stringer("action.type", act.Type),
	)

	answer := &tg.MessagesSetBotCallbackAnswerRequest{
		QueryID:   e.QueryID,
		Message:   fmt.Sprintf("%s(t=%v): %s", e.User.Username, hasToken, act),
		CacheTime: 30,
	}
	if _, err := e.RPC().MessagesSetBotCallbackAnswer(ctx, answer); err != nil {
		return err
	}

	return nil
}
