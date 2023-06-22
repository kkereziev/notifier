package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/joeshaw/envdecode"
)

// Config holds main configuration.
type Config struct {
	Database        Database     `env:""`
	Server          Server       `env:""`
	SlackWebHookURL string       `env:"SLACK_WEB_HOOK_URL" validate:"required"`
	Twilio          TwilioConfig `env:""`
	Mail            MailConfig   `env:""`
}

// New returns a new Config.
func New() (*Config, error) {
	var config Config

	if err := envdecode.Decode(&config); err != nil {
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
