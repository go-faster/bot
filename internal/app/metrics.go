package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/contrib/oteltg"
	"github.com/povilasv/prommod"
	promClient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

type atomicMetric struct {
	atomic.Int64
	promClient.CounterFunc
}

func newMetric(opts promClient.CounterOpts) *atomicMetric {
	m := &atomicMetric{}
	m.CounterFunc = promClient.NewCounterFunc(opts, func() float64 {
		return float64(m.Load())
	})
	return m
}

// Metrics represents bot metrics.
type Metrics struct {
	Start      time.Time
	Messages   *atomicMetric
	Responses  *atomicMetric
	MediaBytes *atomicMetric
	Middleware *oteltg.Middleware

	prometheus     *prometheus.Exporter
	tracerProvider *sdktrace.TracerProvider
	jaeger         *jaeger.Exporter
	resource       *resource.Resource
	mux            *http.ServeMux
	srv            *http.Server
}

// Describe implements prometheus.Collector.
func (m *Metrics) Describe(desc chan<- *promClient.Desc) {
	m.Messages.Describe(desc)
	m.Responses.Describe(desc)
	m.MediaBytes.Describe(desc)
}

// Collect implements prometheus.Collector.
func (m *Metrics) Collect(ch chan<- promClient.Metric) {
	m.Messages.Collect(ch)
	m.Responses.Collect(ch)
	m.MediaBytes.Collect(ch)
}

// Config for metrics.
type Config struct {
	Name      string // app name
	Namespace string // app namespace
	Addr      string // address for metrics server
}

func newPrometheus(config prometheus.Config, options ...controller.Option) (*prometheus.Exporter, error) {
	c := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
		options...,
	)
	return prometheus.New(config, c)
}

func (m *Metrics) registerProfiler() {
	// Routes for pprof.
	m.mux.HandleFunc("/debug/pprof/", pprof.Index)
	m.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	m.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	m.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Manually add support for paths linked to by index page at /debug/pprof/.
	m.mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	m.mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	m.mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	m.mux.Handle("/debug/pprof/block", pprof.Handler("block"))
}

func (m *Metrics) registerPrometheus() {
	// Route for prometheus metrics from registry.
	m.mux.Handle("/metrics", m.prometheus)
}

func (m *Metrics) MeterProvider() metric.MeterProvider {
	return m.prometheus.MeterProvider()
}

func (m *Metrics) TracerProvider() trace.TracerProvider {
	return m.tracerProvider
}

func (m *Metrics) Shutdown(ctx context.Context) error {
	return m.srv.Shutdown(ctx)
}

func (m *Metrics) registerRoot() {
	m.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Briefly describe exported endpoints for admin or devops that has
		// only curl and hope for miracle.
		var b strings.Builder
		b.WriteString("Service is up and running.\n\n")
		b.WriteString("Resource:\n")
		for _, a := range m.resource.Attributes() {
			b.WriteString(fmt.Sprintf("  %-32s %s\n", a.Key, a.Value.AsString()))
		}
		b.WriteString("\nAvailable debug endpoints:\n")
		for _, s := range []struct {
			Name        string
			Description string
		}{
			{"/metrics", "prometheus metrics"},
			{"/debug/pprof", "exported pprof"},
		} {
			b.WriteString(fmt.Sprintf("%-20s - %s\n", s.Name, s.Description))
		}
		_, _ = fmt.Fprintln(w, b.String())
	})
}

const (
	exitCodeOk             = 0
	exitCodeApplicationErr = 1
	exitCodeWatchdog       = 1
)

const (
	shutdownTimeout = time.Second * 5
	watchdogTimeout = shutdownTimeout + time.Second*5
)

const EnvLogLevel = "LOG_LEVEL"

// Run f until interrupt.
func Run(f func(ctx context.Context, log *zap.Logger) error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := zap.NewProductionConfig()

	if s := os.Getenv(EnvLogLevel); s != "" {
		var lvl zapcore.Level
		if err := lvl.UnmarshalText([]byte(s)); err != nil {
			panic(err)
		}
		cfg.Level.SetLevel(lvl)
	}

	lg, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		if err := f(ctx, lg); err != nil {
			return err
		}
		return nil
	})
	go func() {
		// Guaranteed way to kill application.
		<-ctx.Done()

		// Context is canceled, giving application time to shut down gracefully.
		lg.Info("Waiting for application shutdown")
		time.Sleep(watchdogTimeout)

		// Probably deadlock, forcing shutdown.
		lg.Warn("Graceful shutdown watchdog triggered: forcing shutdown")
		os.Exit(exitCodeWatchdog)
	}()

	// Note that we are calling os.Exit() here and no
	if err := wg.Wait(); err != nil {
		lg.Error("Failed",
			zap.Error(err),
		)
		os.Exit(exitCodeApplicationErr)
	}

	os.Exit(exitCodeOk)
}

func (m *Metrics) Run(ctx context.Context) error {
	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		if err := m.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	wg.Go(func() error {
		// Wait until g ctx canceled, then try to shut down server.
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return m.Shutdown(ctx)
	})

	return wg.Wait()
}

// NewMetrics returns new Metrics.
func NewMetrics(log *zap.Logger, cfg Config) (*Metrics, error) {
	res, err := Resource(context.Background(), cfg.Namespace, cfg.Name)
	if err != nil {
		return nil, errors.Wrap(err, "resource")
	}

	registry := promClient.NewPedanticRegistry()
	// Register legacy prometheus-only runtime metrics.
	registry.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
		collectors.NewBuildInfoCollector(),
		prommod.NewCollector(cfg.Name),
	)

	jaegerExporter, err := jaeger.New(jaeger.WithAgentEndpoint())
	if err != nil {
		return nil, errors.Wrap(err, "jaeger")
	}

	promExporter, err := newPrometheus(prometheus.Config{
		DefaultHistogramBoundaries: promClient.DefBuckets,

		Registry:   registry,
		Gatherer:   registry,
		Registerer: registry,
	},
		controller.WithCollectPeriod(0),
		controller.WithResource(res),
	)
	if err != nil {
		return nil, errors.Wrap(err, "prometheus")
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(jaegerExporter),
	)
	mux := http.NewServeMux()
	mw, err := oteltg.New(promExporter.MeterProvider(), tracerProvider)
	if err != nil {
		return nil, errors.Wrap(err, "oteltg")
	}
	m := &Metrics{
		prometheus:     promExporter,
		jaeger:         jaegerExporter,
		tracerProvider: tracerProvider,

		mux: mux,
		srv: &http.Server{
			Handler: mux,
			Addr:    cfg.Addr,
		},

		Middleware: mw,
		Messages: newMetric(promClient.CounterOpts{
			Name: "bot_messages",
			Help: "Total count of received messages",
		}),
		Responses: newMetric(promClient.CounterOpts{
			Name: "bot_responses",
			Help: "Total count of answered messages",
		}),
		MediaBytes: newMetric(promClient.CounterOpts{
			Name: "bot_media_bytes",
			Help: "Total count of received media bytes",
		}),
		Start: time.Now(),
	}

	// Register global OTEL providers.
	global.SetMeterProvider(m.MeterProvider())
	otel.SetTracerProvider(m.tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		),
	)

	m.registerRoot()
	m.registerProfiler()
	m.registerPrometheus()

	log.Info("Metrics initialized",
		zap.Stringer("otel.resource", res),
		zap.String("http.addr", cfg.Addr),
	)

	return m, nil
}

type metricWriter struct {
	Increase func(n int64) int64
	Bytes    int64
}

func (m *metricWriter) Write(p []byte) (n int, err error) {
	delta := int64(len(p))

	m.Increase(delta)
	m.Bytes += delta

	return len(p), nil
}
