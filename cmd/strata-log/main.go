package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ikwukao/strata-log/internal/batcher"
	"github.com/ikwukao/strata-log/internal/config"
	"github.com/ikwukao/strata-log/internal/ingest"
	"github.com/ikwukao/strata-log/internal/pipeline"
	"github.com/ikwukao/strata-log/internal/storage"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer := storage.NewMemoryWriter()

	b, err := batcher.New(
		ctx,
		writer,
		100,
		1*time.Second,
		1000,
	)
	if err != nil {
		logger.Error("failed to initialize batcher", "error", err)
		os.Exit(1)
	}

	processor, err := pipeline.NewLogProcessor(
		ctx,
		4,
		1000,
		func(ctx context.Context, entry ingest.LogEntry) error {
			return b.Submit(entry)
		},
	)
	if err != nil {
		logger.Error("failed to initialize log processor", "error", err)
		b.Close()
		os.Exit(1)
	}

	go func() {
		for err := range b.Errors() {
			logger.Error("batch storage failed", "error", err)
		}
	}()

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

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		cfg.Shutdown.Timeout,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful HTTP shutdown failed", "error", err)
	}

	processor.Close()
	b.Close()

	logger.Info(
		"Strata-Log stopped",
		"stored_logs", writer.Len(),
	)
}
