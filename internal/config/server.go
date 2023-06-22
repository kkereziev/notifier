package config

import "time"

// Server holds metadata for the server of the application.
type Server struct {
	Host            string        `env:"SERVER_HOST" validate:"required"`
	Port            int           `env:"SERVER_PORT" validate:"required"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT,default=20s"`
	KeepAlivePolicy struct {
		Time                time.Duration `env:"GRPC_CLIENT_KEEP_ALIVE_POLICY_TIME,default=10s"`
		Timeout             time.Duration `env:"GRPC_CLIENT_KEEP_ALIVE_POLICY_TIMEOUT,default=30s"`
		PermitWithoutStream bool          `env:"GRPC_CLIENT_KEEP_ALIVE_POLICY_PERMIT_WITHOUT_STREAM,default=true"`
	}
}
