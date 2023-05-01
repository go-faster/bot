package gh

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/zctx"
	"github.com/google/go-github/v52/github"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

func (w *Webhook) clientWithToken(token string) *github.Client {
	return NewTokenClient(token, w.meterProvider, w.tracerProvider)
}

type rateLimitedTransport struct {
	limiter *rate.Limiter
	base    http.RoundTripper
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.limiter.Wait(req.Context()); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}

// NewTokenClient returns new instrumented GitHub client with token and rate limiter.
func NewTokenClient(token string, mp metric.MeterProvider, tp trace.TracerProvider) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	tc.Transport = otelhttp.NewTransport(&rateLimitedTransport{
		limiter: rate.NewLimiter(rate.Every(1*time.Second), 3),
		base:    tc.Transport,
	},
		otelhttp.WithTracerProvider(tp),
		otelhttp.WithMeterProvider(mp),
	)
	return github.NewClient(tc)
}

// Client returns GitHub client for installation.
func (w *Webhook) Client(ctx context.Context) (*github.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tokenKey := fmt.Sprintf("gh:installation:%d:token", w.ghID)
	key, err := w.cache.Get(ctx, tokenKey).Result()
	if err == nil {
		return w.clientWithToken(key), nil
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

	return w.clientWithToken(tok.GetToken()), nil
}
