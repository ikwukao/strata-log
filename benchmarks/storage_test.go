package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/ikwukao/strata-log/internal/ingest"
	"github.com/ikwukao/strata-log/internal/storage"
)

func BenchmarkSQLiteWriteBatch(b *testing.B) {
	writer, err := storage.NewSQLiteWriter(":memory:")
	if err != nil {
		b.Fatalf("create writer: %v", err)
	}
	defer writer.Close()

	entries := make([]ingest.LogEntry, 100)

	for i := range entries {
		entries[i] = ingest.LogEntry{
			Timestamp: time.Now().UTC(),
			Level:     "info",
			Service:   "benchmark",
			Message:   "benchmark log entry",
			Fields: map[string]any{
				"iteration": i,
			},
		}
	}

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := writer.WriteBatch(ctx, entries); err != nil {
			b.Fatalf("write batch: %v", err)
		}
	}
}
