package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-logr/zapr"
	promClient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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
		return metric.NewNoopMeterProvider()
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

const (
	writerStdout = "stdout"
	writerStderr = "stderr"
)

func writerByName(name string) io.Writer {
	switch name {
	case writerStdout:
		return os.Stdout
	case writerStderr:
		return os.Stderr
	default:
		return io.Discard
	}
}

func getEnvOr(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

func newMetrics(ctx context.Context, lg *zap.Logger) (*Metrics, error) {
	otel.SetLogger(zapr.NewLogger(lg.Named("otel")))

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

	// OTEL configuration from environment.
	//
	// See https://opentelemetry.io/docs/concepts/sdk-configuration/general-sdk-configuration/
	const (
		expOTLP       = "otlp"
		expNone       = "none" // no-op
		expPrometheus = "prometheus"
		expJaeger     = "jaeger"

		protoHTTP    = "http"
		protoGRPC    = "grpc"
		defaultProto = protoGRPC
	)

	// Metrics exporter.
	switch exporter := strings.TrimSpace(getEnvOr("OTEL_METRICS_EXPORTER", expOTLP)); exporter {
	case expPrometheus:
		lg.Info("Using prometheus exporter")
		reg := promClient.NewPedanticRegistry()
		exp, err := prometheus.New(
			prometheus.WithRegisterer(reg),
		)
		m.registerShutdown(exporter, exp.Shutdown)
		if err != nil {
			return nil, errors.Wrap(err, "prometheus")
		}
		// Register legacy prometheus-only runtime metrics for backward compatibility.
		reg.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
			collectors.NewBuildInfoCollector(),
		)
		m.prometheus = reg
		m.meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(exp),
		)
	case expOTLP:
		proto := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
		if proto == "" {
			proto = os.Getenv("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL")
		}
		if proto == "" {
			proto = defaultProto
		}
		lg.Info("Using otlp metrics exporter", zap.String("protocol", proto))
		switch proto {
		case protoHTTP:
			exp, err := otlpmetricgrpc.New(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "failed to build grpc trace exporter")
			}
			m.registerShutdown("otlp.metrics.grpc", exp.Shutdown)
			m.meterProvider = sdkmetric.NewMeterProvider(
				sdkmetric.WithResource(res),
				sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
			)
		case protoGRPC:
			exp, err := otlpmetrichttp.New(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "failed to build http trace exporter")
			}
			m.registerShutdown("otlp.metrics.http", exp.Shutdown)
			m.meterProvider = sdkmetric.NewMeterProvider(
				sdkmetric.WithResource(res),
				sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
			)
		default:
			return nil, fmt.Errorf("unsupported metric otlp protocol %q", proto)
		}
	case writerStdout, writerStderr:
		lg.Info(fmt.Sprintf("Using %s periodic metric exporter", exporter))
		enc := json.NewEncoder(writerByName(exporter))
		exp, err := stdoutmetric.New(stdoutmetric.WithEncoder(enc))
		if err != nil {
			return nil, errors.Wrap(err, "stdout metric provider")
		}
		m.registerShutdown(exporter, exp.Shutdown)
		m.meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
		)
	case expNone:
		lg.Info("No metrics exporter is configured by OTEL_METRICS_EXPORTER")
		m.meterProvider = metric.NewNoopMeterProvider()
	default:
		return nil, errors.Errorf("unsupported OTEL_METRICS_EXPORTER %q", exporter)
	}

	// Traces exporter.
	switch exporter := strings.TrimSpace(getEnvOr("OTEL_TRACES_EXPORTER", expOTLP)); exporter {
	case expJaeger:
		lg.Info("Using jaeger exporter")
		var jaegerOptions []jaeger.AgentEndpointOption
		jaegerOptions = append(jaegerOptions,
			jaeger.WithLogger(zap.NewStdLog(lg.Named("jaeger"))),
		)
		exp, err := jaeger.New(jaeger.WithAgentEndpoint(jaegerOptions...))
		if err != nil {
			return nil, errors.Wrap(err, "jaeger")
		}
		m.registerShutdown(exporter, exp.Shutdown)
		m.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(exp),
		)
	case expOTLP:
		proto := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
		if proto == "" {
			proto = os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL")
		}
		if proto == "" {
			proto = defaultProto
		}
		lg.Info("Using otlp traces exporter", zap.String("protocol", proto))
		switch proto {
		case protoGRPC:
			exp, err := otlptracegrpc.New(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create trace exporter: %w", err)
			}
			m.registerShutdown("otlp.traces", exp.Shutdown)
			m.tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithResource(res),
				sdktrace.WithBatcher(exp),
			)
		case protoHTTP:
			exp, err := otlptracehttp.New(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create trace exporter: %w", err)
			}
			m.registerShutdown("otlp.traces", exp.Shutdown)
			m.tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithResource(res),
				sdktrace.WithBatcher(exp),
			)
		default:
			return nil, fmt.Errorf("unsupported traces otlp protocol %q", proto)
		}
	case writerStdout, writerStderr:
		lg.Info(fmt.Sprintf("Using %s traces exporter", exporter))
		exp, err := stdouttrace.New(stdouttrace.WithWriter(writerByName(exporter)))
		if err != nil {
			return nil, errors.Wrap(err, "stdout")
		}
		m.registerShutdown(exporter, exp.Shutdown)
		m.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(exp),
		)
	case expNone:
		lg.Info("No traces exporter is configured by OTEL_TRACES_EXPORTER")
		m.tracerProvider = trace.NewNoopTracerProvider()
	default:
		return nil, errors.Errorf("unsupported OTEL_TRACES_EXPORTER %q", exporter)
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
