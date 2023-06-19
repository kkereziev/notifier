package internal

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joeshaw/envdecode"
)

// Config holds the configuration for the program.
type Config struct {
	Server          Server       `env:""`
	SlackWebHookURL string       `env:"SLACK_WEB_HOOK_URL" validate:"required"`
	Retry           RequestRetry `env:""`
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

// Server holds the configuration for the HTTP server.
type Server struct {
	Host            string        `env:"SERVER_HOST" validate:"required"`
	Port            int           `env:"SERVER_PORT" validate:"required"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT,default=5s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT,default=10s"`
	IdleTimeout     time.Duration `env:"SERVER_IDLE_TIMEOUT,default=120s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT,default=20s"`
}

// Addr retrieves the address of the server.
func (s Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// RequestRetry holds the config for retrying requests.
type RequestRetry struct {
	MaxRetries int           `env:"MAX_RETRIES,default=3"`
	Delay      time.Duration `env:"MAX_DELAY,default=5s"`
}
