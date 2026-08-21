// Strata-Log is a high-performance structured logging service.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ikwukao/strata-log/internal/config"
	"github.com/ikwukao/strata-log/internal/ingest"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", ingest.HealthHandler)
	mux.HandleFunc("/v1/logs", ingest.LogHandler)

	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	go func() {
		logger.Info(
			"starting Strata-Log",
			"address", server.Addr,
			"storage", cfg.Storage.Path,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-stop

	logger.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		cfg.Shutdown.Timeout,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Strata-Log stopped")
}
