package botapi

import (
	"net/http"
)

// Options is Client options.
type Options struct {
	HTTPClient *http.Client
}

func (m *Options) setDefaults() {
	if m.HTTPClient == nil {
		m.HTTPClient = http.DefaultClient
	}
}
