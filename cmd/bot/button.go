package main

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v50/github"
	"github.com/gotd/td/tg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/action"
	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/ent/user"
)

func (a *App) OnButton(ctx context.Context, e dispatch.Button) (rerr error) {
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
	rpc := e.RPC()
	defer func() {
		if rerr != nil {
			span.SetStatus(codes.Error, rerr.Error())
			answer := &tg.MessagesSetBotCallbackAnswerRequest{
				QueryID: e.QueryID,
				Message: fmt.Sprintf("Error: %s", rerr),
				Alert:   true,
			}
			if _, err := rpc.MessagesSetBotCallbackAnswer(ctx, answer); err != nil {
				rerr = multierr.Append(rerr, err)
			}
		}
	}()

	var act action.Action
	if err := act.UnmarshalText(e.Data); err != nil {
		return errors.Wrap(err, "unmarshal")
	}
	span.SetAttributes(
		attribute.Int("action.id", act.ID),
		attribute.Int64("action.repository_id", act.RepositoryID),
		attribute.Stringer("action.entity", act.Entity),
		attribute.Stringer("action.type", act.Type),
	)

	var token string
	{
		users, err := a.db.User.Query().Where(
			user.ID(e.User.ID),
		).All(ctx)
		if err != nil {
			return errors.Wrap(err, "query user")
		}
		for _, u := range users {
			if u.GithubToken != "" {
				token = u.GithubToken
				break
			}
		}
	}
	if token == "" {
		return errors.New("no PAT token found for user")
	}

	switch {
	case act.Is(action.Merge, action.PullRequest):
		api := a.clientWithToken(ctx, token)
		repo, _, err := api.Repositories.GetByID(ctx, act.RepositoryID)
		if err != nil {
			return errors.Wrap(err, "get repo")
		}
		var (
			owner    = repo.GetOwner().GetLogin()
			repoName = repo.GetName()
			message  = "" // use default message
			options  = &github.PullRequestOptions{
				MergeMethod: "merge",
			}
		)
		if _, _, err := api.PullRequests.Merge(ctx, owner, repoName, act.ID, message, options); err != nil {
			return errors.Wrap(err, "merge")
		}
		if _, err := rpc.MessagesSetBotCallbackAnswer(ctx, &tg.MessagesSetBotCallbackAnswerRequest{
			QueryID: e.QueryID,
			Message: "Pull request merged",
		}); err != nil {
			return errors.Wrap(err, "answer")
		}
		return nil
	default:
		return errors.Errorf("unknown action %q(%q)", act.Type, act.Entity)
	}
}
