package app

import (
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/go-faster/sdk/profiler"
)

func (m *Metrics) registerProfiler() {
	var routes []string
	if v := os.Getenv("PPROF_ROUTES"); v != "" {
		routes = strings.Split(v, ",")
	}
	if len(routes) == 1 && routes[0] == "none" {
		return
	}
	opt := profiler.Options{
		Routes: routes,
		UnknownRoute: func(route string) {
			m.lg.Warn("Unknown pprof route", zap.String("route", route))
		},
	}
	m.mux.Handle("/debug/pprof/", profiler.New(opt))
}
