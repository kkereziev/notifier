package config

import "fmt"

// Database holds metadata for the database of the application.
type Database struct {
	Host     string `env:"DATABASE_HOST" validate:"required"`
	Port     int    `env:"DATABASE_PORT" validate:"required"`
	User     string `env:"DATABASE_USER" validate:"required"`
	Password string `env:"DATABASE_PASSWORD" validate:"required"`
	Database string `env:"DATABASE_NAME" validate:"required"`
}

// DSN returns the database source name.
func (d Database) DSN() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
	)
}
