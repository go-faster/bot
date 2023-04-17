package main

import (
	"encoding/base64"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v50/github"
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
		return nil, errors.Wrap(err, "create github transport")
	}
	return github.NewClient(&http.Client{
		Transport: ghTransport,
	}), nil
}
