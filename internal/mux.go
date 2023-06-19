package internal

import (
	"github.com/dimfeld/httptreemux/v5"
)

const (
	_apiURLPattern    = "/api/v1"
	_slackEndpointURL = "/slack"
)

// Notifier is manages notification sending.
//
// Implementations of Notifier must be save for concurrent use by multiple goroutines.
type Notifier interface {
	SlackNotifier
}

// NewMux is a constructor function for creating new multiplexer for the HTTP server.
func NewMux(config *Config, notifier Notifier) *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()

	registerRoutes(config, mux, notifier)

	return mux
}

func registerRoutes(config *Config, m *httptreemux.ContextMux, notifier Notifier) {
	g := m.NewGroup(_apiURLPattern)

	g.POST(_slackEndpointURL, MakeSlackEndpoint(config, notifier))
}
