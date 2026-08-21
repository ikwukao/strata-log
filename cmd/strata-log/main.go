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
	"github.com/ikwukao/strata-log/internal/pipeline"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	processor, err := pipeline.NewLogProcessor(
		context.Background(),
		4,
		1000,
		func(ctx context.Context, entry ingest.LogEntry) error {
			logger.Info(
				"log received",
				"timestamp", entry.Timestamp,
				"level", entry.Level,
				"service", entry.Service,
				"message", entry.Message,
			)

			return nil
		},
	)
	if err != nil {
		logger.Error("failed to initialize log processor", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", ingest.HealthHandler)
	mux.Handle("/v1/logs", ingest.ProcessHandler(processor))

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
		logger.Error("graceful HTTP shutdown failed", "error", err)
		processor.Close()
		os.Exit(1)
	}

	processor.Close()

	logger.Info("Strata-Log stopped")
}
