package gpt

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterMap[K comparable] struct {
	limiters map[K]*rate.Limiter
	create   func(K) *rate.Limiter
	mux      sync.Mutex
}

func newLimiterMap[K comparable](create func(K) *rate.Limiter) *limiterMap[K] {
	return &limiterMap[K]{
		limiters: map[K]*rate.Limiter{},
		create:   create,
	}
}

func (m *limiterMap[K]) Allow(k K) (time.Duration, bool) {
	// Limit disabled.
	if m == nil {
		return 0, true
	}

	m.mux.Lock()
	defer m.mux.Unlock()

	if m.limiters == nil {
		m.limiters = map[K]*rate.Limiter{}
	}

	limiter, ok := m.limiters[k]
	if !ok {
		limiter = m.create(k)
		m.limiters[k] = limiter
	}

	r := limiter.Reserve()
	return r.Delay(), r.OK()
}
