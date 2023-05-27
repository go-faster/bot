package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/brpaz/echozap"
	"github.com/go-faster/errors"
	sdkapp "github.com/go-faster/sdk/app"
	"github.com/go-faster/sdk/zctx"
	"github.com/google/uuid"
	"github.com/gotd/contrib/oteltg"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/go-faster/bot/internal/api"
	"github.com/go-faster/bot/internal/app"
	"github.com/go-faster/bot/internal/botapi"
	"github.com/go-faster/bot/internal/dispatch"
	"github.com/go-faster/bot/internal/ent"
	"github.com/go-faster/bot/internal/entdb"
	"github.com/go-faster/bot/internal/entgaps"
	"github.com/go-faster/bot/internal/entsession"
	"github.com/go-faster/bot/internal/gh"
	"github.com/go-faster/bot/internal/oas"
	"github.com/go-faster/bot/internal/otelredis"
)

type App struct {
	client     *telegram.Client
	token      string
	raw        *tg.Client
	sender     *message.Sender
	dispatcher tg.UpdateDispatcher
	mux        *dispatch.MessageMux
	tracer     trace.Tracer
	openai     *openai.Client
	m          *sdkapp.Metrics
	lg         *zap.Logger
	wh         *gh.Webhook
	db         *ent.Client
	cache      *redis.Client
	gaps       *updates.Manager
	rdy        *Readiness
}

func initApp(m *sdkapp.Metrics, lg *zap.Logger) (_ *App, rerr error) {
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

	db, err := entdb.Open(os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, errors.Wrap(err, "open database")
	}
	gapsStore := entgaps.New(db, m.TracerProvider())
	dispatcher := tg.NewUpdateDispatcher()
	updatesHandler := updates.New(updates.Config{
		Handler:        dispatcher,
		Storage:        gapsStore,
		Logger:         lg.Named("gaps"),
		TracerProvider: m.TracerProvider(),
	})

	otg, err := oteltg.New(m.MeterProvider(), m.TracerProvider())
	if err != nil {
		return nil, errors.Wrap(err, "otel")
	}

	uuidNameSpaceBotToken := uuid.MustParse("24085c34-5e70-4b1b-9fd9-a82a98879839")
	client := telegram.NewClient(appID, appHash, telegram.Options{
		UpdateHandler: updatesHandler,
		Logger:        lg.Named("client"),
		SessionStorage: entsession.Storage{
			Database: db,
			UUID:     uuid.NewSHA1(uuidNameSpaceBotToken, []byte(token)),
			Tracer:   m.TracerProvider().Tracer("entsession"),
		},
		Middlewares: []telegram.Middleware{
			telegram.MiddlewareFunc(func(next tg.Invoker) telegram.InvokeFunc {
				return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
					return next.Invoke(zctx.Base(ctx, lg), input, output)
				}
			}),
			// NB: This is critical for updates handler to work.
			updhook.UpdateHook(updatesHandler.Handle),
			otg,
		},
	})
	raw := client.API()
	sender := message.NewSender(raw)
	dd := downloader.NewDownloader()
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
		return nil, errors.Wrap(err, "setup github installation")
	}
	ghInstallationID, err := strconv.ParseInt(os.Getenv("GITHUB_INSTALLATION_ID"), 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "GITHUB_INSTALLATION_ID")
	}
	mux := dispatch.NewMessageMux().
		WithTracerProvider(m.TracerProvider())

	webhook := gh.NewWebhook(
		db,
		ghInstallationClient,
		ghInstallationID,
		sender,
		m.MeterProvider(),
		m.TracerProvider(),
	).WithCache(r)
	if notifyGroup, ok := os.LookupEnv("TG_NOTIFY_GROUP"); ok {
		webhook = webhook.WithNotifyGroup(notifyGroup)
	}
	if secret := os.Getenv("GITHUB_SECRET"); secret != "" {
		webhook = webhook.WithSecret(secret)
	}

	openaiConfig := openai.DefaultConfig(os.Getenv("OPENAI_TOKEN"))
	openaiConfig.HTTPClient = &http.Client{
		Transport: httpTransport,
		Timeout:   time.Minute,
	}

	a := &App{
		cache:      r,
		db:         db,
		client:     client,
		gaps:       updatesHandler,
		token:      token,
		raw:        raw,
		sender:     sender,
		dispatcher: dispatcher,
		mux:        mux,
		m:          m,
		lg:         lg,
		wh:         webhook,
		openai:     openai.NewClientWithConfig(openaiConfig),
		rdy:        new(Readiness),
		tracer:     m.TracerProvider().Tracer(""),
	}

	var h dispatch.MessageHandler = app.NewMiddleware(mux, dd, m, app.MiddlewareOptions{
		BotAPI: botapi.NewClient(token, botapi.Options{
			HTTPClient: &http.Client{
				Transport: httpTransport,
				Timeout:   15 * time.Second,
			},
		}),
		Logger: lg.Named("metrics"),
	})
	_ = dispatch.NewBot(raw).
		WithSender(sender).
		WithTracerProvider(m.TracerProvider()).
		Register(dispatcher).
		OnMessage(gh.NewHook(h, db.LastChannelMessage)).
		OnButton(a)

	return a, nil
}

func (a *App) Close() error {
	return a.db.Close()
}

// Readiness is a simple readiness check aggregator.
type Readiness struct {
	registry []*atomic.Bool
}

func (r *Readiness) Ready() bool {
	for _, v := range r.registry {
		if !v.Load() {
			return false
		}
	}
	return true
}

func (r *Readiness) Register() *atomic.Bool {
	v := new(atomic.Bool)
	r.registry = append(r.registry, v)
	return v
}

func (a *App) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	if a.wh.HasSecret() {
		lg := a.lg.Named("webhook")

		httpAddr := os.Getenv("HTTP_BOT_ADDR")
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

			// Pass logger.
			func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					req := c.Request()
					logCtx := zctx.Base(req.Context(), lg.Named("handler"))
					req = req.WithContext(logCtx)
					c.SetRequest(req)
					return next(c)
				}
			},
		)
		e.GET("/probe/startup", func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})
		e.GET("/probe/ready", func(c echo.Context) error {
			if !a.rdy.Ready() {
				return c.String(http.StatusServiceUnavailable, "not ready")
			}
			return c.String(http.StatusOK, "ok")
		})
		a.wh.RegisterRoutes(e)
		server := http.Server{
			Addr:    httpAddr,
			Handler: e,
		}
		g.Go(func() error {
			if err := a.wh.Run(ctx); err != nil {
				return errors.Wrap(err, "webhook task")
			}
			return nil
		})
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
	{
		// API Server.
		lg := a.lg.Named("api")
		httpAddr := os.Getenv("HTTP_API_ADDR")
		if httpAddr == "" {
			httpAddr = "localhost:8080"
		}
		h, err := oas.NewServer(api.NewServer(a.db),
			oas.WithMeterProvider(a.m.MeterProvider()),
			oas.WithTracerProvider(a.m.TracerProvider()),
		)
		if err != nil {
			return errors.Wrap(err, "oas server")
		}
		server := http.Server{
			Addr: httpAddr,
			Handler: otelhttp.NewHandler(cors.Default().Handler(h), "",
				otelhttp.WithMeterProvider(a.m.MeterProvider()),
				otelhttp.WithTracerProvider(a.m.TracerProvider()),
				otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
					op, ok := h.FindRoute(r.Method, r.URL.Path)
					if ok {
						return "http." + op.OperationID()
					}
					return operation
				}),
			),
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
		ticker := time.NewTicker(time.Minute * 2)
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
	readyRedis := a.rdy.Register()
	g.Go(func() error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		if _, err := a.cache.Ping(ctx).Result(); err != nil {
			return errors.Wrap(err, "ping redis")
		}

		readyRedis.Store(true)
		a.lg.Info("Redis connection established")
		return nil
	})
	readyClickhouse := a.rdy.Register()
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
		readyClickhouse.Store(true)
		if err := db.Close(); err != nil {
			return errors.Wrap(err, "close clickhouse")
		}
		return nil
	})
	readyTelegram := a.rdy.Register()
	g.Go(func() error {
		return a.client.Run(ctx, func(ctx context.Context) error {
			a.lg.Debug("client initialized")

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
				a.lg.Info("Bot logged in",
					zap.String("name", status.User.Username),
				)
			} else {
				a.lg.Info("Bot login restored",
					zap.String("name", status.User.Username),
				)
			}

			ready := make(chan struct{})
			g, ctx := errgroup.WithContext(ctx)
			g.Go(func() error {
				self, err := a.client.Self(ctx)
				if err != nil {
					return errors.Wrap(err, "self")
				}
				if err := a.gaps.Run(ctx, a.client.API(), self.ID, updates.AuthOptions{
					IsBot: true,
					OnStart: func(ctx context.Context) {
						close(ready)
					},
				}); err != nil {
					return errors.Wrap(err, "gaps")
				}
				return nil
			})
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ready:
				}

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
				readyTelegram.Store(true)
				return nil
			})
			return g.Wait()
		})
	})
	return g.Wait()
}

func runBot(ctx context.Context, m *sdkapp.Metrics, lg *zap.Logger) (rerr error) {
	a, err := initApp(m, lg)
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
