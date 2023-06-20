package internal

import (
	"context"
	"log"

	"github.com/dimfeld/httptreemux/v5"
)

const (
	_apiURLPattern    = "/api/v1"
	_slackEndpointURL = "/slack"
	_smsEndpointURL   = "/sms"
	_mailEndpointURL  = "/mail"
)

// SlackNotifier manages sending of notification via Slack.
type SlackNotifier interface {
	NotifySlack(context.Context, any) error
}

// SMSNotifier manages sending of notification via SMS.
type SMSNotifier interface {
	NotifySMS(context.Context, any) error
}

// MailNotifier manages sending of notification via mail.
type MailNotifier interface {
	NotifyMail(context.Context, any) error
}

// Notifier is manages notification sending.
//
// Implementations of Notifier must be save for concurrent use by multiple goroutines.
//
//go:generate moq -out mocks/notifier.go -pkg mocks . Notifier
type Notifier interface {
	SlackNotifier
	SMSNotifier
	MailNotifier
}

// NewMux is a constructor function for creating new multiplexer for the HTTP server.
func NewMux(config *Config, logger *log.Logger, notifier Notifier) *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()

	registerRoutes(config, logger, mux, notifier)

	return mux
}

func registerRoutes(config *Config, logger *log.Logger, m *httptreemux.ContextMux, notifier Notifier) {
	g := m.NewGroup(_apiURLPattern)

	g.Use(CORSMiddleware)
	g.Use(RecoverMiddleware(logger))
	g.Use(LoggingMiddleware(logger))

	g.POST(_slackEndpointURL, MakeSlackEndpoint(config, notifier))
	g.POST(_smsEndpointURL, MakeSMSEndpoint(config, notifier))
	g.POST(_mailEndpointURL, MakeMailEndpoint(config, notifier))
}
