package commits

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v52/github"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"github.com/go-faster/bot/internal/app"
	"github.com/go-faster/bot/internal/entdb"
	"github.com/go-faster/bot/internal/otelredis"
	"github.com/go-faster/bot/internal/stat"
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

func Root() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commits",
		Short: "Gather commit information and save to database",
		Run: func(cmd *cobra.Command, args []string) {
			app.Run(func(ctx context.Context, logger *zap.Logger, m *app.Metrics) error {
				tracer := m.TracerProvider().Tracer("command")

				ctx, span := tracer.Start(ctx, "job.commits")
				defer span.End()

				httpTransport := otelhttp.NewTransport(http.DefaultTransport,
					otelhttp.WithTracerProvider(m.TracerProvider()),
					otelhttp.WithMeterProvider(m.MeterProvider()),
				)

				r := redis.NewClient(&redis.Options{
					Addr: "redis:6379",
				})
				r.AddHook(otelredis.NewHook(m.TracerProvider()))

				ghInstallationClient, err := setupGithubInstallation(httpTransport)
				if err != nil {
					return errors.Wrap(err, "setup github installation")
				}
				ghInstallationID, err := strconv.ParseInt(os.Getenv("GITHUB_INSTALLATION_ID"), 10, 64)
				if err != nil {
					return errors.Wrap(err, "GITHUB_INSTALLATION_ID")
				}
				db, err := entdb.Open(os.Getenv("DATABASE_URL"))
				if err != nil {
					return errors.Wrap(err, "open database")
				}

				c := stat.NewCommit(db, r, ghInstallationClient, ghInstallationID, m.MeterProvider(), m.TracerProvider())
				if err := c.Update(ctx); err != nil {
					return errors.Wrap(err, "update commits")
				}

				return nil
			})
		},
	}
	return cmd
}
