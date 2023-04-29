package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/sdk/autometer"
	"github.com/go-faster/sdk/autotracer"
	"github.com/go-logr/zapr"
	promClient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Metrics implement common basic metrics and infrastructure to it.
type Metrics struct {
	lg *zap.Logger

	prometheus *promClient.Registry

	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider

	resource   *resource.Resource
	mux        *http.ServeMux
	srv        *http.Server
	propagator propagation.TextMapPropagator

	shutdowns []shutdown
}

func (m *Metrics) registerShutdown(name string, fn func(ctx context.Context) error) {
	m.shutdowns = append(m.shutdowns, shutdown{name: name, fn: fn})
}

type shutdown struct {
	name string
	fn   func(ctx context.Context) error
}

func (m *Metrics) String() string {
	return "metrics"
}

func (m *Metrics) run(ctx context.Context) error {
	defer m.lg.Debug("Stopped metrics")
	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		m.lg.Info("Starting metrics server")
		if err := m.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		m.lg.Debug("Metrics server gracefully stopped")
		return nil
	})
	wg.Go(func() error {
		// Wait until g ctx canceled, then try to shut down server.
		<-ctx.Done()

		m.lg.Debug("Shutting down metrics")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return m.shutdown(ctx)
	})

	return wg.Wait()
}

func (m *Metrics) shutdown(ctx context.Context) error {
	var (
		wg   sync.WaitGroup
		l    sync.Mutex
		errs []error
	)

	// Launch shutdowns in parallel.
	wg.Add(len(m.shutdowns))
	for _, s := range m.shutdowns {
		var (
			f = s.fn
			n = s.name
		)
		go func() {
			defer wg.Done()
			if err := f(ctx); err != nil {
				e := errors.Wrapf(err, "shutdown %s", n)
				l.Lock()
				errs = append(errs, e)
				l.Unlock()
			}
		}()
	}

	// Wait for all shutdowns to finish.
	wg.Wait()

	// Combine all shutdown errors.
	l.Lock()
	err := multierr.Combine(errs...)
	l.Unlock()

	return err
}

func (m *Metrics) registerPrometheus() {
	// Route for prometheus metrics from registry.
	m.mux.Handle("/metrics",
		promhttp.HandlerFor(m.prometheus, promhttp.HandlerOpts{}),
	)
}

func (m *Metrics) MeterProvider() metric.MeterProvider {
	if m.meterProvider == nil {
		return global.MeterProvider()
	}
	return m.meterProvider
}

func (m *Metrics) TracerProvider() trace.TracerProvider {
	if m.tracerProvider == nil {
		return trace.NewNoopTracerProvider()
	}
	return m.tracerProvider
}

func (m *Metrics) TextMapPropagator() propagation.TextMapPropagator {
	return m.propagator
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
		type Endpoint struct {
			Path        string
			Description string
		}
		endpoints := []Endpoint{
			{"/debug/pprof", "exported pprof"},
		}
		if m.prometheus != nil {
			endpoints = append(endpoints, Endpoint{
				Path:        "/metrics",
				Description: "exported prometheus metrics",
			})
		}
		for _, s := range endpoints {
			b.WriteString(fmt.Sprintf("%-20s - %s\n", s.Path, s.Description))
		}
		_, _ = fmt.Fprintln(w, b.String())
	})
}

func prometheusAddr() string {
	host := "localhost"
	port := "9464"
	if v := os.Getenv("OTEL_EXPORTER_PROMETHEUS_HOST"); v != "" {
		host = v
	}
	if v := os.Getenv("OTEL_EXPORTER_PROMETHEUS_PORT"); v != "" {
		port = v
	}
	return net.JoinHostPort(host, port)
}

type zapErrorHandler struct {
	lg *zap.Logger
}

func (z zapErrorHandler) Handle(err error) {
	z.lg.Error("Error", zap.Error(err))
}

func newMetrics(ctx context.Context, lg *zap.Logger) (*Metrics, error) {
	{
		// Setup global OTEL logger and error handler.
		logger := lg.Named("otel")
		otel.SetLogger(zapr.NewLogger(logger))
		otel.SetErrorHandler(zapErrorHandler{lg: logger})
	}
	addr := prometheusAddr()
	if v := os.Getenv("METRICS_ADDR"); v != "" {
		addr = v
	}
	res, err := Resource(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "resource")
	}

	mux := http.NewServeMux()
	m := &Metrics{
		lg:       lg,
		resource: res,
		mux:      mux,
		srv: &http.Server{
			Handler:           mux,
			Addr:              addr,
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       time.Second,
			WriteTimeout:      time.Second,
		},
	}

	m.registerShutdown("http", m.srv.Shutdown)
	{
		provider, stop, err := autotracer.NewTracerProvider(ctx, autotracer.WithResource(res))
		if err != nil {
			return nil, errors.Wrap(err, "tracer provider")
		}
		m.tracerProvider = provider
		m.registerShutdown("tracer", stop)
	}
	{
		provider, stop, err := autometer.NewMeterProvider(ctx,
			autometer.WithResource(res),
			autometer.WithOnPrometheusRegistry(func(reg *promClient.Registry) {
				m.prometheus = reg
			}),
		)
		if err != nil {
			return nil, errors.Wrap(err, "meter provider")
		}
		m.meterProvider = provider
		m.registerShutdown("meter", stop)
	}

	// Automatically composited from the OTEL_PROPAGATORS environment variable.
	m.propagator = autoprop.NewTextMapPropagator()

	// Setting up go runtime metrics.
	if err := runtime.Start(
		runtime.WithMeterProvider(m.MeterProvider()),
		runtime.WithMinimumReadMemStatsInterval(time.Second), // export as env?
	); err != nil {
		return nil, errors.Wrap(err, "runtime metrics")
	}

	// Register global OTEL providers.
	global.SetMeterProvider(m.MeterProvider())
	otel.SetTracerProvider(m.TracerProvider())
	otel.SetTextMapPropagator(m.TextMapPropagator())

	// Register basic http routes.
	m.registerRoot()
	m.registerProfiler()
	if m.prometheus != nil {
		m.registerPrometheus()
	}

	lg.Info("Metrics initialized",
		zap.Stringer("otel.resource", res),
		zap.String("metrics.http.addr", addr),
	)

	return m, nil
}
