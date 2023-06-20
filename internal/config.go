package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joeshaw/envdecode"
)

// Config holds the configuration for the program.
type Config struct {
	Server          ServerConfig       `env:""`
	SlackWebHookURL string             `env:"SLACK_WEB_HOOK_URL" validate:"required"`
	Retry           RequestRetryConfig `env:""`
	Twilio          TwilioConfig       `env:""`
	Mail            MailConfig         `env:""`
}

// NewConfig is a constructor function for Config.
func NewConfig() (*Config, error) {
	var config Config

	err := envdecode.Decode(&config)
	if err != nil {
		if err != envdecode.ErrNoTargetFieldsAreSet {
			return nil, err
		}
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) validate() error {
	return validator.New().Struct(c)
}

// ServerConfig holds the configuration for the HTTP server.
type ServerConfig struct {
	Host            string        `env:"SERVER_HOST" validate:"required"`
	Port            int           `env:"SERVER_PORT" validate:"required"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT,default=5s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT,default=10s"`
	IdleTimeout     time.Duration `env:"SERVER_IDLE_TIMEOUT,default=120s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT,default=20s"`
}

// Addr retrieves the address of the server.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// RequestRetryConfig holds the config for retrying requests.
type RequestRetryConfig struct {
	MaxRetries int           `env:"MAX_RETRIES,default=3"`
	Delay      time.Duration `env:"MAX_DELAY,default=2s"`
}

// TwilioConfig holds configuration for Twilio service.
type TwilioConfig struct {
	SID    string `env:"TWILIO_SID" validate:"required"`
	Token  string `env:"TWILIO_TOKEN" validate:"required"`
	Number string `env:"TWILIO_NUMBER" validate:"required"`
}

// MailConfig holds configuration for Mail config.
type MailConfig struct {
	EmailSender  string `env:"EMAIL_SENDER" validate:"required"`
	SMTPHost     string `env:"EMAIL_SMTP_HOST" validate:"required"`
	SMTPPort     int    `env:"EMAIL_SMTP_PORT,default=456" validate:"required"`
	SMTPUsername string `env:"EMAIL_SMTP_USERNAME" validate:"required"`
	SMTPPassword string `env:"EMAIL_SMTP_PASSWORD" validate:"required"`
}
