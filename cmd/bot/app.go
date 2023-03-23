package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/brpaz/echozap"
	"github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v45/github"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	bolt "go.etcd.io/bbolt"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"

	"github.com/go-faster/bot/internal/app"
	"github.com/go-faster/bot/internal/botapi"
	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/gh"
	"github.com/go-faster/bot/internal/storage"
)

type App struct {
	client *telegram.Client
	token  string
	raw    *tg.Client
	sender *message.Sender

	stateStorage *BoltState
	gaps         *updates.Manager
	dispatcher   tg.UpdateDispatcher

	db      *pebble.DB
	storage storage.MsgID
	mux     dispatch.MessageMux
	bot     *dispatch.Bot

	github *github.Client
	http   *http.Client
	m      *app.Metrics
	lg     *zap.Logger
	wh     *gh.Webhook
}

func InitApp(ctx context.Context, m *app.Metrics, lg *zap.Logger) (_ *App, rerr error) {
	// Reading app id from env (never hardcode it!).
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		return nil, errors.Wrapf(err, "APP_ID not set or invalid %q", os.Getenv("APP_ID"))
	}

	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		return nil, errors.New("no APP_HASH provided")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		return nil, errors.New("no BOT_TOKEN provided")
	}

	// Setting up session storage.
	cacheDir := "/cache"
	sessionDir := filepath.Join(cacheDir, ".td")
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return nil, errors.Wrap(err, "mkdir")
	}

	stateDb, err := bolt.Open(filepath.Join(sessionDir, "gaps-state.bbolt"), fs.ModePerm, bolt.DefaultOptions)
	if err != nil {
		return nil, errors.Wrap(err, "state database")
	}
	defer func() {
		if rerr != nil {
			multierr.AppendInto(&rerr, stateDb.Close())
		}
	}()

	db, err := pebble.Open(
		filepath.Join(sessionDir, fmt.Sprintf("bot.%s.state", tokHash(token))),
		&pebble.Options{},
	)
	if err != nil {
		return nil, errors.Wrap(err, "database")
	}
	defer func() {
		if rerr != nil {
			multierr.AppendInto(&rerr, db.Close())
		}
	}()
	msgIDStore := storage.NewMsgID(db)

	stateStorage := NewBoltState(stateDb)
	dispatcher := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: dispatcher,
		Storage: stateStorage,
		Logger:  lg.Named("gaps"),
	})
	client := telegram.NewClient(appID, appHash, telegram.Options{
		Logger: lg.Named("client"),
		SessionStorage: &session.FileStorage{
			Path: filepath.Join(sessionDir, sessionFileName(token)),
		},
		UpdateHandler: dispatch.NewLoggedDispatcher(
			gaps, lg.Named("updates"), m.TracerProvider(),
		),
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(func(ctx context.Context, u tg.UpdatesClass) error {
				go func() {
					if err := gaps.Handle(ctx, u); err != nil {
						lg.Error("Handle RPC response update error", zap.Error(err))
					}
				}()
				return nil
			}),
		},
	})
	raw := client.API()
	sender := message.NewSender(raw)
	dd := downloader.NewDownloader()
	httpTransport := otelhttp.NewTransport(http.DefaultTransport,
		otelhttp.WithTracerProvider(m.TracerProvider()),
		otelhttp.WithMeterProvider(m.MeterProvider()),
	)
	httpClient := &http.Client{
		Transport: httpTransport,
		Timeout:   15 * time.Second,
	}

	mux := dispatch.NewMessageMux()
	var h dispatch.MessageHandler = app.NewMiddleware(mux, dd, m, app.MiddlewareOptions{
		BotAPI: botapi.NewClient(token, botapi.Options{
			HTTPClient: httpClient,
		}),
		Logger: lg.Named("metrics"),
	})
	h = storage.NewHook(h, msgIDStore)

	b := dispatch.NewBot(raw).
		WithSender(sender).
		WithLogger(lg).
		Register(dispatcher).
		OnMessage(h)

	webhook := gh.NewWebhook(msgIDStore, sender, m.MeterProvider(), m.TracerProvider()).
		WithLogger(lg)
	if notifyGroup, ok := os.LookupEnv("TG_NOTIFY_GROUP"); ok {
		webhook = webhook.WithNotifyGroup(notifyGroup)
	}
	if secret := os.Getenv("GITHUB_SECRET"); secret != "" {
		webhook = webhook.WithSecret(secret)
	}

	a := &App{
		client:       client,
		token:        token,
		raw:          raw,
		sender:       sender,
		stateStorage: stateStorage,
		gaps:         gaps,
		dispatcher:   dispatcher,
		db:           db,
		storage:      msgIDStore,
		mux:          mux,
		bot:          b,
		http:         httpClient,
		m:            m,
		lg:           lg,
		wh:           webhook,
	}

	if v, ok := os.LookupEnv("GITHUB_APP_ID"); ok {
		ghClient, err := setupGithub(v, httpTransport)
		if err != nil {
			return nil, errors.Wrap(err, "setup github")
		}
		a.github = ghClient
	}

	return a, nil
}

func (a *App) Close() error {
	return multierr.Append(a.stateStorage.db.Close(), a.db.Close())
}

func (a *App) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	if a.wh.HasSecret() {
		lg := a.lg.Named("webhook")

		httpAddr := os.Getenv("HTTP_ADDR")
		if httpAddr == "" {
			httpAddr = "localhost:8080"
		}
		e := echo.New()
		e.Use(
			middleware.Recover(),
			middleware.RequestID(),
			echozap.ZapLogger(lg.Named("requests")),
			otelecho.Middleware("bot",
				otelecho.WithTracerProvider(a.m.TracerProvider()),
			),
		)
		e.GET("/status", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})
		a.wh.RegisterRoutes(e)
		server := http.Server{
			Addr:    httpAddr,
			Handler: e,
		}
		g.Go(func() error {
			lg.Info("ListenAndServe", zap.String("addr", server.Addr))
			return server.ListenAndServe()
		})
		g.Go(func() error {
			<-ctx.Done()
			shutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			lg.Info("Shutdown", zap.String("addr", server.Addr))
			if err := server.Shutdown(shutCtx); err != nil {
				return multierr.Append(err, server.Close())
			}
			return nil
		})
	}
	g.Go(func() error {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case now := <-ticker.C:
				if err := a.FetchEvents(ctx, now.Add(-time.Minute*10)); err != nil {
					a.lg.Error("FetchEvents error", zap.Error(err))
				}
			}
		}
	})
	g.Go(func() error {
		rdb := redis.NewClient(&redis.Options{
			Addr: "redis:6379",
		})

		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		if _, err := rdb.Ping(ctx).Result(); err != nil {
			return errors.Wrap(err, "ping redis")
		}

		a.lg.Info("Redis connection established")
		return nil
	})
	g.Go(func() error {
		db, err := ch.Dial(ctx, ch.Options{
			Address:        os.Getenv("CLICKHOUSE_ADDR"),
			Compression:    ch.CompressionZSTD,
			TracerProvider: a.m.TracerProvider(),
			MeterProvider:  a.m.MeterProvider(),
			Database:       "faster",

			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
			User:     os.Getenv("CLICKHOUSE_USER"),

			OpenTelemetryInstrumentation: true,
		})
		if err != nil {
			return errors.Wrap(err, "connect to clickhouse")
		}
		a.lg.Info("Clickhouse connection established",
			zap.Stringer("server", db.ServerInfo()),
		)
		if err := db.Ping(ctx); err != nil {
			return errors.Wrap(err, "ping clickhouse")
		}
		if err := db.Close(); err != nil {
			return errors.Wrap(err, "close clickhouse")
		}
		return nil
	})
	g.Go(func() error {
		return a.client.Run(ctx, func(ctx context.Context) error {
			a.lg.Debug("Client initialized")

			au := a.client.Auth()
			status, err := au.Status(ctx)
			if err != nil {
				return errors.Wrap(err, "auth status")
			}

			if !status.Authorized {
				if _, err := au.Bot(ctx, a.token); err != nil {
					return errors.Wrap(err, "login")
				}

				// Refresh auth status.
				status, err = au.Status(ctx)
				if err != nil {
					return errors.Wrap(err, "auth status")
				}
			} else {
				a.lg.Info("Bot login restored",
					zap.String("name", status.User.Username),
				)
			}

			if err := a.gaps.Auth(ctx, a.raw, status.User.ID, status.User.Bot, false); err != nil {
				return err
			}
			defer func() { _ = a.gaps.Logout() }()

			if _, disableRegister := os.LookupEnv("DISABLE_COMMAND_REGISTER"); !disableRegister {
				if err := a.mux.RegisterCommands(ctx, a.raw); err != nil {
					return errors.Wrap(err, "register commands")
				}
			}

			if deployNotify := os.Getenv("TG_DEPLOY_NOTIFY_GROUP"); deployNotify != "" {
				p, err := a.sender.ResolveDomain(deployNotify, peer.OnlyChannel).AsInputPeer(ctx)
				if err != nil {
					return errors.Wrap(err, "resolve")
				}
				info, _ := debug.ReadBuildInfo()

				var commit string
				for _, c := range info.Settings {
					switch c.Key {
					case "vcs.revision":
						commit = c.Value[:7]
					}
				}

				var options []message.StyledTextOption
				options = append(options,
					styling.Plain("🚀 Started "),
					styling.Italic(fmt.Sprintf("(%s, gotd %s, layer: %d) ",
						info.GoVersion, app.GetGotdVersion(), tg.Layer),
					),
					styling.Code(commit),
				)

				var mrkp tg.ReplyMarkupClass
				if module := info.Main.Path; module != "" && strings.HasPrefix(module, "github.com") {
					commitLink := fmt.Sprintf("https://%s/commit/%s", module, commit)
					mrkp = markup.InlineRow(markup.URL("Commit", commitLink))
				}

				if _, err := a.sender.To(p).Markup(mrkp).StyledText(ctx, options...); err != nil {
					return errors.Wrap(err, "send")
				}
			}

			<-ctx.Done()
			return ctx.Err()
		})
	})
	return g.Wait()
}

func runBot(ctx context.Context, m *app.Metrics, lg *zap.Logger) (rerr error) {
	a, err := InitApp(ctx, m, lg)
	if err != nil {
		return errors.Wrap(err, "initialize")
	}
	defer func() {
		multierr.AppendInto(&rerr, a.Close())
	}()

	if err := setupBot(a); err != nil {
		return errors.Wrap(err, "setup")
	}

	if err := a.Run(ctx); err != nil {
		return errors.Wrap(err, "run")
	}
	return nil
}
