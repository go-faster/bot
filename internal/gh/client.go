package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v50/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func (w *Webhook) clientWithToken(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// Client returns GitHub client for installation.
func (w *Webhook) Client(ctx context.Context) (*github.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tokenKey := fmt.Sprintf("gh:installation:%d:token", w.ghID)
	key, err := w.cache.Get(ctx, tokenKey).Result()
	if err == nil {
		return w.clientWithToken(ctx, key), nil
	}

	tok, _, err := w.ghClient.Apps.CreateInstallationToken(ctx, w.ghID, &github.InstallationTokenOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "create token")
	}

	expiration := time.Until(tok.GetExpiresAt().Time)
	zctx.From(ctx).Info("Token expires in",
		zap.Duration("d", expiration),
	)
	offset := time.Minute * 10
	if expiration > offset {
		// Just to make sure that we will not get expired token.
		expiration -= offset
	}
	if _, err := w.cache.Set(ctx, tokenKey, tok.GetToken(), expiration).Result(); err != nil {
		return nil, errors.Wrap(err, "set token")
	}

	return w.clientWithToken(ctx, tok.GetToken()), nil
}
