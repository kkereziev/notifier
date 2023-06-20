package internal

import (
	"context"

	"github.com/dimfeld/httptreemux/v5"
)

const (
	_apiURLPattern    = "/api/v1"
	_slackEndpointURL = "/slack"
)

// SlackNotifier manages sending of notification via Slack.
type SlackNotifier interface {
	NotifySlack(context.Context, any) error
}

// Notifier is manages notification sending.
//
// Implementations of Notifier must be save for concurrent use by multiple goroutines.
//
//go:generate moq -out mocks/notifier.go -pkg mocks . Notifier
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
