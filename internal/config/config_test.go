package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("STRATA_LOG_HOST", "")
	t.Setenv("STRATA_LOG_PORT", "")
	t.Setenv("STRATA_LOG_STORAGE_PATH", "")
	t.Setenv("STRATA_LOG_BUFFER_CAPACITY", "")
	t.Setenv("STRATA_LOG_BATCH_SIZE", "")
	t.Setenv("STRATA_LOG_FLUSH_PERIOD", "")
	t.Setenv("STRATA_LOG_PIPELINE_WORKERS", "")
	t.Setenv("STRATA_LOG_SHUTDOWN_TIMEOUT", "")

	cfg := Load()

	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf(
			"expected default host 0.0.0.0, got %q",
			cfg.Server.Host,
		)
	}

	if cfg.Server.Port != 9090 {
		t.Fatalf(
			"expected default port 9090, got %d",
			cfg.Server.Port,
		)
	}

	if cfg.Storage.Path != "strata-log.db" {
		t.Fatalf(
			"expected default storage path strata-log.db, got %q",
			cfg.Storage.Path,
		)
	}

	if cfg.Buffer.Capacity != 10_000 {
		t.Fatalf(
			"expected default buffer capacity 10000, got %d",
			cfg.Buffer.Capacity,
		)
	}

	if cfg.Batcher.MaxSize != 100 {
		t.Fatalf(
			"expected default batch size 100, got %d",
			cfg.Batcher.MaxSize,
		)
	}

	if cfg.Batcher.FlushPeriod != 500*time.Millisecond {
		t.Fatalf(
			"expected default flush period 500ms, got %s",
			cfg.Batcher.FlushPeriod,
		)
	}

	if cfg.Pipeline.Workers != 4 {
		t.Fatalf(
			"expected default pipeline workers 4, got %d",
			cfg.Pipeline.Workers,
		)
	}

	if cfg.Shutdown.Timeout != 10*time.Second {
		t.Fatalf(
			"expected default shutdown timeout 10s, got %s",
			cfg.Shutdown.Timeout,
		)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("STRATA_LOG_HOST", "127.0.0.1")
	t.Setenv("STRATA_LOG_PORT", "9090")
	t.Setenv("STRATA_LOG_STORAGE_PATH", "/tmp/strata-log.db")
	t.Setenv("STRATA_LOG_BUFFER_CAPACITY", "5000")
	t.Setenv("STRATA_LOG_BATCH_SIZE", "250")
	t.Setenv("STRATA_LOG_FLUSH_PERIOD", "1s")
	t.Setenv("STRATA_LOG_PIPELINE_WORKERS", "8")
	t.Setenv("STRATA_LOG_SHUTDOWN_TIMEOUT", "20s")

	cfg := Load()

	if cfg.Server.Host != "127.0.0.1" {
		t.Fatalf(
			"expected host 127.0.0.1, got %q",
			cfg.Server.Host,
		)
	}

	if cfg.Server.Port != 9090 {
		t.Fatalf(
			"expected port 9090, got %d",
			cfg.Server.Port,
		)
	}

	if cfg.Storage.Path != "/tmp/strata-log.db" {
		t.Fatalf(
			"expected custom storage path, got %q",
			cfg.Storage.Path,
		)
	}

	if cfg.Buffer.Capacity != 5000 {
		t.Fatalf(
			"expected buffer capacity 5000, got %d",
			cfg.Buffer.Capacity,
		)
	}

	if cfg.Batcher.MaxSize != 250 {
		t.Fatalf(
			"expected batch size 250, got %d",
			cfg.Batcher.MaxSize,
		)
	}

	if cfg.Batcher.FlushPeriod != time.Second {
		t.Fatalf(
			"expected flush period 1s, got %s",
			cfg.Batcher.FlushPeriod,
		)
	}

	if cfg.Pipeline.Workers != 8 {
		t.Fatalf(
			"expected pipeline workers 8, got %d",
			cfg.Pipeline.Workers,
		)
	}

	if cfg.Shutdown.Timeout != 20*time.Second {
		t.Fatalf(
			"expected shutdown timeout 20s, got %s",
			cfg.Shutdown.Timeout,
		)
	}
}
