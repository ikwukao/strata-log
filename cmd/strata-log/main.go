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
	"github.com/ikwukao/strata-log/internal/query"
	"github.com/ikwukao/strata-log/internal/storage"
	"github.com/ikwukao/strata-log/internal/telemetry"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	metrics := telemetry.NewMetrics()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the SQLite storage backend.
	writer, err := storage.NewSQLiteWriter(cfg.Storage.Path)
	if err != nil {
		logger.Error(
			"failed to initialize storage",
			"error", err,
			"path", cfg.Storage.Path,
		)
		os.Exit(1)
	}
	defer writer.Close()

	// SQLiteWriter implements both storage.Reader and storage.Writer.
	var reader storage.Reader = writer

	// Initialize the query service used by GET /v1/logs.
	queryService, err := query.NewService(reader)
	if err != nil {
		logger.Error(
			"failed to initialize query service",
			"error", err,
		)
		os.Exit(1)
	}

	queryHandler := query.NewHandler(queryService)

	// Initialize the batcher responsible for grouping log entries
	// before they are persisted.
	b, err := batcher.New(
		ctx,
		writer,
		100,
		1*time.Second,
		1000,
	)
	if err != nil {
		logger.Error(
			"failed to initialize batcher",
			"error", err,
		)
		os.Exit(1)
	}

	// Initialize the asynchronous log processor.
	processor, err := pipeline.NewLogProcessor(
		ctx,
		4,
		1000,
		func(ctx context.Context, entry ingest.LogEntry) error {
			return b.Submit(entry)
		},
	)
	if err != nil {
		logger.Error(
			"failed to initialize log processor",
			"error", err,
		)
		b.Close()
		os.Exit(1)
	}

	// Create the ingestion handler once and reuse it for every request.
	ingestHandler := ingest.ProcessHandler(processor)

	// Monitor asynchronous storage failures from the batcher.
	go func() {
		for err := range b.Errors() {
			metrics.IncErrors()

			logger.Error(
				"batch storage failed",
				"error", err,
			)
		}
	}()

	mux := http.NewServeMux()

	// Health endpoint.
	mux.HandleFunc(
		"/healthz",
		ingest.HealthHandler,
	)

	// Metrics endpoint for Prometheus-compatible monitoring.
	mux.HandleFunc(
		"/metrics",
		func(w http.ResponseWriter, r *http.Request) {
			metrics.ServeHTTP(w, r)
		},
	)

	// POST /v1/logs -> asynchronous log ingestion.
	// GET  /v1/logs -> persisted log queries.
	mux.HandleFunc(
		"/v1/logs",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				ingestHandler(w, r)

			case http.MethodGet:
				queryHandler.ServeHTTP(w, r)

			default:
				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
			}
		},
	)

	// Strata-Log listens on 9090 by default.
	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer signal.Stop(stop)

	go func() {
		logger.Info(
			"starting Strata-Log",
			"address", server.Addr,
			"storage", cfg.Storage.Path,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"server failed",
				"error", err,
			)
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
		logger.Error(
			"graceful HTTP shutdown failed",
			"error", err,
		)
	}

	// Stop accepting new log entries and flush pending batches.
	processor.Close()
	b.Close()

	if err := writer.Close(); err != nil {
		logger.Error(
			"storage close failed",
			"error", err,
		)
	}

	logger.Info(
		"Strata-Log stopped",
		"storage", cfg.Storage.Path,
	)
}
