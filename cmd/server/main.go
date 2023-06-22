package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kkereziev/notifier/v2/internal"
	"github.com/kkereziev/notifier/v2/internal/adding"
	"github.com/kkereziev/notifier/v2/internal/config"
	"github.com/kkereziev/notifier/v2/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	log, err := initLog("notifier")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//nolint:errcheck
	defer log.Sync()

	if err := run(log); err != nil {
		log.Fatal("startup", "SIGTERM", err)
	}
}

func initLog(serviceName string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
	}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}

	//nolint:errcheck
	defer log.Sync()

	return log.Sugar(), nil
}

func run(log *zap.SugaredLogger) error {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := storage.NewConnection(cfg)
	if err != nil {
		log.Fatal(err)
	}

	idempotencyRepo := storage.NewIdempotencyStorage(conn)
	notificationRepo := storage.NewNotificationRepository(conn)

	server, err := internal.NewInstance(cfg, log, adding.NewService(log, conn, idempotencyRepo, notificationRepo), nil)
	if err != nil {
		log.Fatal(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- server.Start(context.Background())
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown completed", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		server.Stop(ctx)
	}

	return nil
}
