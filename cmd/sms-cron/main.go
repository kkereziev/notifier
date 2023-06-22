package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kkereziev/notifier/internal/config"
	"github.com/kkereziev/notifier/internal/notifying"
	"github.com/kkereziev/notifier/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	log, err := initLog("notifier")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//nolint:errcheck
	defer log.Sync()

	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := storage.NewConnection(cfg)
	if err != nil {
		log.Fatal(err)
	}

	notificationRepo := storage.NewNotificationRepository(conn)

	s := notifying.NewService(cfg, log, notificationRepo, conn)

	for {
		if err := s.NotifySMS(context.Background()); err != nil {
			return err
		}

		time.Sleep(time.Second * 10)
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
