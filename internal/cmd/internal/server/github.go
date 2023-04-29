package server

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
)

func setupGithubInstallation(httpTransport http.RoundTripper) (*github.Client, error) {
	ghAppID, err := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "GITHUB_APP_ID is invalid")
	}
	key, err := base64.StdEncoding.DecodeString(os.Getenv("GITHUB_PRIVATE_KEY"))
	if err != nil {
		return nil, errors.Wrap(err, "GITHUB_PRIVATE_KEY is invalid")
	}
	ghTransport, err := ghinstallation.NewAppsTransport(httpTransport, ghAppID, key)
	if err != nil {
		return nil, errors.Wrap(err, "create ghInstallation transport")
	}
	return github.NewClient(&http.Client{
		Transport: ghTransport,
	}), nil
}

func (a *App) clientWithToken(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
