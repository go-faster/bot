package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gotd/td/tg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/action"
	"github.com/go-faster/bot/internal/dispatch"
)

func (a *App) OnButton(ctx context.Context, e dispatch.Button) error {
	ctx, span := a.tracer.Start(ctx, "OnBotCallbackQuery")
	defer span.End()
	if e.Input == nil {
		span.SetStatus(codes.Ok, "Ignored")
		a.lg.Info("OnButton: no user")
		return nil
	}
	a.lg.Info("OnButton",
		zap.String("user", e.Input.String()),
	)

	var act action.Action
	if err := json.Unmarshal(e.Data, &act); err != nil {
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

	span.SetAttributes(
		attribute.Int("action.id", act.ID),
		attribute.String("action.entity", act.Entity),
		attribute.String("action.action", act.Action),
	)

	answer := &tg.MessagesSetBotCallbackAnswerRequest{
		QueryID:   e.QueryID,
		Message:   fmt.Sprintf("Hello, %s!", e.User.FirstName),
		CacheTime: 30,
	}
	if _, err := e.RPC().MessagesSetBotCallbackAnswer(ctx, answer); err != nil {
		return err
	}

	return nil
}
