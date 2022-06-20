package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"

	"github.com/gotd/contrib/middleware/tg_prom"
)

type metric struct {
	atomic.Int64
	prometheus.CounterFunc
}

func newMetric(opts prometheus.CounterOpts) *metric {
	m := &metric{}
	m.CounterFunc = prometheus.NewCounterFunc(opts, func() float64 {
		return float64(m.Load())
	})
	return m
}

// Metrics represents bot metrics.
type Metrics struct {
	Start      time.Time
	Messages   *metric
	Responses  *metric
	MediaBytes *metric
	Middleware *tg_prom.Middleware
}

// Describe implements prometheus.Collector.
func (m Metrics) Describe(desc chan<- *prometheus.Desc) {
	m.Messages.Describe(desc)
	m.Responses.Describe(desc)
	m.MediaBytes.Describe(desc)
	for _, mm := range m.Middleware.Metrics() {
		mm.Describe(desc)
	}
}

// Collect implements prometheus.Collector.
func (m Metrics) Collect(ch chan<- prometheus.Metric) {
	m.Messages.Collect(ch)
	m.Responses.Collect(ch)
	m.MediaBytes.Collect(ch)
	for _, mm := range m.Middleware.Metrics() {
		mm.Collect(ch)
	}
}

// NewMetrics returns new Metrics.
func NewMetrics() Metrics {
	return Metrics{
		Middleware: tg_prom.New(),
		Messages: newMetric(prometheus.CounterOpts{
			Name: "bot_messages",
			Help: "Total count of received messages",
		}),
		Responses: newMetric(prometheus.CounterOpts{
			Name: "bot_responses",
			Help: "Total count of answered messages",
		}),
		MediaBytes: newMetric(prometheus.CounterOpts{
			Name: "bot_media_bytes",
			Help: "Total count of received media bytes",
		}),
		Start: time.Now(),
	}
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
