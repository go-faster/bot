package main

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

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
