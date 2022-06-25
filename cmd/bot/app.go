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
	"time"

	"github.com/brpaz/echozap"
	"github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"
	"github.com/google/go-github/v42/github"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	bolt "go.etcd.io/bbolt"
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

	"github.com/go-faster/bot/internal/botapi"
	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/gh"
	"github.com/go-faster/bot/internal/metrics"
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
	mts    metrics.Metrics
	logger *zap.Logger
}

func InitApp(mts metrics.Metrics, logger *zap.Logger) (_ *App, rerr error) {
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
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "get home")
	}
	sessionDir := filepath.Join(home, ".td")
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
		Logger:  logger.Named("gaps"),
	})
	client := telegram.NewClient(appID, appHash, telegram.Options{
		Logger: logger.Named("client"),
		SessionStorage: &session.FileStorage{
			Path: filepath.Join(sessionDir, sessionFileName(token)),
		},
		UpdateHandler: dispatch.NewLoggedDispatcher(
			gaps, logger.Named("updates"),
		),
		Middlewares: []telegram.Middleware{
			// mts.Middleware, // HACK(ernado): fix contrib
			updhook.UpdateHook(func(ctx context.Context, u tg.UpdatesClass) error {
				go func() {
					if err := gaps.Handle(ctx, u); err != nil {
						logger.Error("Handle RPC response update error", zap.Error(err))
					}
				}()
				return nil
			}),
		},
	})
	raw := client.API()
	sender := message.NewSender(raw)
	dd := downloader.NewDownloader()
	httpTransport := http.DefaultTransport
	httpClient := &http.Client{
		Transport: httpTransport,
		Timeout:   15 * time.Second,
	}

	mux := dispatch.NewMessageMux()
	var h dispatch.MessageHandler = metrics.NewMiddleware(mux, dd, mts, metrics.MiddlewareOptions{
		BotAPI: botapi.NewClient(token, botapi.Options{
			HTTPClient: httpClient,
		}),
		Logger: logger.Named("metrics"),
	})
	h = storage.NewHook(h, msgIDStore)

	b := dispatch.NewBot(raw).
		WithSender(sender).
		WithLogger(logger).
		Register(dispatcher).
		OnMessage(h)

	app := &App{
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
		mts:          mts,
		logger:       logger,
	}

	if v, ok := os.LookupEnv("GITHUB_APP_ID"); ok {
		ghClient, err := setupGithub(v, httpTransport)
		if err != nil {
			return nil, errors.Wrap(err, "setup github")
		}
		app.github = ghClient
	}

	return app, nil
}

func (b *App) Close() error {
	return multierr.Append(b.stateStorage.db.Close(), b.db.Close())
}

func (b *App) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	if secret, ok := os.LookupEnv("GITHUB_SECRET"); ok {
		logger := b.logger.Named("webhook")

		httpAddr := os.Getenv("HTTP_ADDR")
		if httpAddr == "" {
			httpAddr = "localhost:8080"
		}

		webhook := gh.NewWebhook(b.storage, b.sender, secret).
			WithLogger(logger)
		if notifyGroup, ok := os.LookupEnv("TG_NOTIFY_GROUP"); ok {
			webhook = webhook.WithNotifyGroup(notifyGroup)
		}

		e := echo.New()
		e.Use(
			middleware.Recover(),
			middleware.RequestID(),
			echozap.ZapLogger(logger.Named("requests")),
		)

		e.GET("/status", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})
		webhook.RegisterRoutes(e)

		server := http.Server{
			Addr:    httpAddr,
			Handler: e,
		}
		group.Go(func() error {
			logger.Info("ListenAndServe", zap.String("addr", server.Addr))
			return server.ListenAndServe()
		})
		group.Go(func() error {
			<-ctx.Done()
			shutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			logger.Info("Shutdown", zap.String("addr", server.Addr))
			if err := server.Shutdown(shutCtx); err != nil {
				return multierr.Append(err, server.Close())
			}
			return nil
		})
	}

	group.Go(func() error {
		return b.client.Run(ctx, func(ctx context.Context) error {
			b.logger.Debug("Client initialized")

			au := b.client.Auth()
			status, err := au.Status(ctx)
			if err != nil {
				return errors.Wrap(err, "auth status")
			}

			if !status.Authorized {
				if _, err := au.Bot(ctx, b.token); err != nil {
					return errors.Wrap(err, "login")
				}

				// Refresh auth status.
				status, err = au.Status(ctx)
				if err != nil {
					return errors.Wrap(err, "auth status")
				}
			} else {
				b.logger.Info("Bot login restored",
					zap.String("name", status.User.Username),
				)
			}

			if err := b.gaps.Auth(ctx, b.raw, status.User.ID, status.User.Bot, false); err != nil {
				return err
			}
			defer func() { _ = b.gaps.Logout() }()

			if _, disableRegister := os.LookupEnv("DISABLE_COMMAND_REGISTER"); !disableRegister {
				if err := b.mux.RegisterCommands(ctx, b.raw); err != nil {
					return errors.Wrap(err, "register commands")
				}
			}

			if deployNotify := os.Getenv("TG_DEPLOY_NOTIFY_GROUP"); deployNotify != "" {
				p, err := b.sender.ResolveDomain(deployNotify, peer.OnlyChannel).AsInputPeer(ctx)
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
					styling.Italic(fmt.Sprintf("(%s, %s, layer: %d) ",
						info.GoVersion, metrics.GetVersion(), tg.Layer),
					),
					styling.Code(commit),
				)
				if _, err := b.sender.To(p).StyledText(ctx, options...); err != nil {
					return errors.Wrap(err, "send")
				}
			}

			<-ctx.Done()
			return ctx.Err()
		})
	})
	return group.Wait()
}

func runBot(ctx context.Context, mts metrics.Metrics, logger *zap.Logger) (rerr error) {
	app, err := InitApp(mts, logger)
	if err != nil {
		return errors.Wrap(err, "initialize")
	}
	defer func() {
		multierr.AppendInto(&rerr, app.Close())
	}()

	if err := setupBot(app); err != nil {
		return errors.Wrap(err, "setup")
	}

	if err := app.Run(ctx); err != nil {
		return errors.Wrap(err, "run")
	}
	return nil
}
