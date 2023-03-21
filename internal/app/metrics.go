package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	rpprof "runtime/pprof"
	"strings"

	"github.com/go-faster/errors"
	promClient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	jaegerp "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
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

		m.lg.Debug("Shutting down metrics server")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return m.shutdown(ctx)
	})

	return wg.Wait()
}

func (m *Metrics) shutdown(ctx context.Context) error {
	return m.srv.Shutdown(ctx)
}

func (m *Metrics) registerProfiler() {
	var routes []string
	if v := os.Getenv("GO_PPROF_ROUTES"); v != "" {
		routes = strings.Split(v, ",")
	}
	if len(routes) == 1 && routes[0] == "none" {
		return
	}
	if len(routes) == 0 {
		// Enable all routes by default except cmdline (unsafe).
		//
		// Route name is "/debug/pprof/<name>".
		routes = []string{
			// From pprof.<Name>.
			"profile",
			"symbol",
			"trace",

			// From pprof.Handler(<name>).
			"goroutine",
			"heap",
			"threadcreate",
			"block",
		}
	}
	m.lg.Info("Registering pprof routes", zap.Strings("routes", routes))
	m.mux.HandleFunc("/debug/pprof/", pprof.Index)
	for _, name := range routes {
		route := path.Join("/debug/pprof/", name)
		switch name {
		case "cmdline":
			m.mux.HandleFunc(route, pprof.Cmdline)
		case "profile":
			m.mux.HandleFunc(route, pprof.Profile)
		case "symbol":
			m.mux.HandleFunc(route, pprof.Symbol)
		case "trace":
			m.mux.HandleFunc(route, pprof.Trace)
		case "none": // invalid
			m.lg.Warn("Invalid pprof route ('none' should be the only one route specified)",
				zap.String("route", name),
			)
		default:
			if rpprof.Lookup(name) == nil {
				m.lg.Warn("Invalid pprof route", zap.String("route", name))
				continue
			}
			m.mux.Handle(route, pprof.Handler(name))
		}
	}
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

func writerByName(name string) io.Writer {
	switch name {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	default:
		return io.Discard
	}
}

func newMetrics(ctx context.Context, lg *zap.Logger) (*Metrics, error) {
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
			Handler: mux,
			Addr:    addr,
		},
	}

	// OTEL configuration from environment.
	//
	// See https://opentelemetry.io/docs/concepts/sdk-configuration/general-sdk-configuration/

	// Metrics exporter.
	switch exporter := os.Getenv("OTEL_METRICS_EXPORTER"); exporter {
	case "prometheus":
		lg.Info("Using prometheus exporter")
		registry := promClient.NewPedanticRegistry()
		promExporter, err := prometheus.New(
			prometheus.WithRegisterer(registry),
		)
		if err != nil {
			return nil, errors.Wrap(err, "prometheus")
		}
		// Register legacy prometheus-only runtime metrics for backward compatibility.
		registry.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
			collectors.NewBuildInfoCollector(),
		)
		m.prometheus = registry
		m.meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(promExporter),
		)
	case "stdout", "stderr":
		lg.Info(fmt.Sprintf("Using %s periodic metric exporter", exporter))
		enc := json.NewEncoder(writerByName(exporter))
		exp, err := stdoutmetric.New(stdoutmetric.WithEncoder(enc))
		if err != nil {
			return nil, errors.Wrap(err, "stdout metric provider")
		}
		m.meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
		)
	case "", "none":
		lg.Info("No metrics exporter is configured by OTEL_METRICS_EXPORTER")
		m.meterProvider = metric.NewNoopMeterProvider()
	default:
		return nil, errors.Errorf("unsupported OTEL_METRICS_EXPORTER %q", exporter)
	}

	// Traces exporter.
	switch exporter := os.Getenv("OTEL_TRACES_EXPORTER"); exporter {
	case "jaeger":
		lg.Info("Using jaeger exporter")
		var jaegerOptions []jaeger.AgentEndpointOption
		jaegerOptions = append(jaegerOptions,
			jaeger.WithLogger(zap.NewStdLog(lg.Named("jaeger"))),
		)
		jaegerExporter, err := jaeger.New(jaeger.WithAgentEndpoint(jaegerOptions...))
		if err != nil {
			return nil, errors.Wrap(err, "jaeger")
		}
		m.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(jaegerExporter),
		)
	case "stdout", "stderr":
		lg.Info(fmt.Sprintf("Using %s traces exporter", exporter))
		stdoutExporter, err := stdouttrace.New(stdouttrace.WithWriter(writerByName(exporter)))
		if err != nil {
			return nil, errors.Wrap(err, "stdout")
		}
		m.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(stdoutExporter),
		)
	case "none", "":
		lg.Info("No traces exporter is configured by OTEL_TRACES_EXPORTER")
		m.tracerProvider = trace.NewNoopTracerProvider()
	default:
		return nil, errors.Errorf("unsupported OTEL_TRACES_EXPORTER %q", exporter)
	}

	// Propagators.
	propagators := "tracecontext,baggage" // default as per OTEL convention
	if v := os.Getenv("OTEL_PROPAGATORS"); v != "" {
		propagators = v
	}
	if propagators == "none" {
		m.propagator = propagation.NewCompositeTextMapPropagator() // noop
		m.lg.Info("Propagation is disabled by OTEL_PROPAGATORS")
	} else {
		var (
			list           []propagation.TextMapPropagator
			valid, invalid []string
		)
		for _, p := range strings.Split(propagators, ",") {
			// See https://opentelemetry.io/docs/concepts/sdk-configuration/general-sdk-configuration/#otel_propagators
			switch p {
			case "tracecontext":
				list = append(list, propagation.TraceContext{})
			case "baggage":
				list = append(list, propagation.Baggage{})
			case "jaeger":
				list = append(list, jaegerp.Jaeger{})
			// TODO(ernado): support b3, b3multi?
			default:
				invalid = append(invalid, p)
				continue
			}
			valid = append(valid, p)
		}
		m.propagator = propagation.NewCompositeTextMapPropagator(list...)
		if len(valid) > 0 {
			m.lg.Info("Propagators configured", zap.Strings("propagators", valid))
		} else {
			m.lg.Info("No propagators configured")
		}
		if len(invalid) > 0 {
			m.lg.Warn("Unsupported propagators", zap.Strings("propagators.invalid", invalid))
		}
	}

	// TODO: Register OTEL runtime metrics.

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
