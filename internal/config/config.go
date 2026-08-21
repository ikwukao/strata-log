package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Storage  StorageConfig
	Buffer   BufferConfig
	Batcher  BatcherConfig
	Shutdown ShutdownConfig
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func (c ServerConfig) Address() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

type StorageConfig struct {
	Path string
}

type BufferConfig struct {
	Capacity int
}

type BatcherConfig struct {
	MaxSize     int
	FlushPeriod time.Duration
}

type ShutdownConfig struct {
	Timeout time.Duration
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Host:            getEnv("STRATA_LOG_HOST", "0.0.0.0"),
			Port:            getEnvInt("STRATA_LOG_PORT", 9090),
			ReadTimeout:     getEnvDuration("STRATA_LOG_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvDuration("STRATA_LOG_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     getEnvDuration("STRATA_LOG_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvDuration("STRATA_LOG_SHUTDOWN_TIMEOUT", 10*time.Second),
		},

		Storage: StorageConfig{
			Path: getEnv("STRATA_LOG_STORAGE_PATH", "strata-log.db"),
		},

		Buffer: BufferConfig{
			Capacity: getEnvInt("STRATA_LOG_BUFFER_CAPACITY", 10_000),
		},

		Batcher: BatcherConfig{
			MaxSize:     getEnvInt("STRATA_LOG_BATCH_SIZE", 100),
			FlushPeriod: getEnvDuration("STRATA_LOG_FLUSH_PERIOD", 500*time.Millisecond),
		},

		Shutdown: ShutdownConfig{
			Timeout: getEnvDuration("STRATA_LOG_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
	}
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
