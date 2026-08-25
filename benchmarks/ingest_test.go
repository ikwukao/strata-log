package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/ikwukao/strata-log/internal/ingest"
	"github.com/ikwukao/strata-log/internal/pipeline"
)

func BenchmarkLogPipeline(b *testing.B) {
	ctx := context.Background()

	processor, err := pipeline.NewLogProcessor(
		ctx,
		4,
		b.N,
		func(context.Context, ingest.LogEntry) error {
			return nil
		},
	)
	if err != nil {
		b.Fatalf("create processor: %v", err)
	}

	defer processor.Close()

	entry := ingest.LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     "info",
		Service:   "benchmark",
		Message:   "benchmark log entry",
		Fields: map[string]any{
			"environment": "benchmark",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := processor.Submit(entry); err != nil {
			b.Fatalf("submit entry: %v", err)
		}
	}
}
