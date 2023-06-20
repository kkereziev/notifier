package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kkereziev/notifier/internal"
)

const _serviceName = "notifier "

func main() {
	log := log.New(os.Stdout, _serviceName, log.LstdFlags)

	if err := run(log); err != nil {
		log.Fatalf("failed startup: %v", err)
	}
}

func run(log *log.Logger) error {
	cfg, err := internal.NewConfig()
	if err != nil {
		return fmt.Errorf("config initialization: %v", err)
	}

	s := internal.NewService(cfg)

	server := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      internal.NewMux(cfg, log, s),
		IdleTimeout:  cfg.Server.IdleTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("[Server] listen on %s", server.Addr)

		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Println("shutdown started", sig)
		defer log.Println("shutdown completed", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			//nolint: errcheck
			server.Close()

			return fmt.Errorf("could not stop server gracefully: %v", err)
		}
	}

	return nil
}
