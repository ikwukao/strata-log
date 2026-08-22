package query

import (
	"context"
	"errors"
	"testing"

	"github.com/ikwukao/strata-log/internal/storage"
)

type mockReader struct {
	records []storage.LogRecord
	options storage.QueryOptions
	err     error
}

func (m *mockReader) QueryLogs(
	_ context.Context,
	options storage.QueryOptions,
) ([]storage.LogRecord, error) {
	m.options = options

	if m.err != nil {
		return nil, m.err
	}

	return m.records, nil
}

func TestNewServiceRequiresReader(t *testing.T) {
	_, err := NewService(nil)

	if err == nil {
		t.Fatal("expected error when reader is nil")
	}
}

func TestServiceQueryUsesDefaultLimit(t *testing.T) {
	reader := &mockReader{
		records: []storage.LogRecord{
			{
				ID:      1,
				Message: "test log",
			},
		},
	}

	service, err := NewService(reader)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	records, err := service.Query(
		context.Background(),
		storage.QueryOptions{},
	)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if reader.options.Limit != DefaultLimit {
		t.Fatalf(
			"expected default limit %d, got %d",
			DefaultLimit,
			reader.options.Limit,
		)
	}
}

func TestServiceQueryCapsLimit(t *testing.T) {
	reader := &mockReader{}

	service, err := NewService(reader)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.Query(
		context.Background(),
		storage.QueryOptions{
			Limit: 5000,
		},
	)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	if reader.options.Limit != MaxLimit {
		t.Fatalf(
			"expected max limit %d, got %d",
			MaxLimit,
			reader.options.Limit,
		)
	}
}

func TestServiceQueryPreservesFilters(t *testing.T) {
	reader := &mockReader{}

	service, err := NewService(reader)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	options := storage.QueryOptions{
		Level:   "error",
		Service: "payment-api",
		Limit:   25,
	}

	_, err = service.Query(
		context.Background(),
		options,
	)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}

	if reader.options.Level != "error" {
		t.Fatalf(
			"expected level %q, got %q",
			"error",
			reader.options.Level,
		)
	}

	if reader.options.Service != "payment-api" {
		t.Fatalf(
			"expected service %q, got %q",
			"payment-api",
			reader.options.Service,
		)
	}

	if reader.options.Limit != 25 {
		t.Fatalf(
			"expected limit 25, got %d",
			reader.options.Limit,
		)
	}
}

func TestServiceQueryPropagatesReaderError(t *testing.T) {
	expectedErr := errors.New("storage unavailable")

	reader := &mockReader{
		err: expectedErr,
	}

	service, err := NewService(reader)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.Query(
		context.Background(),
		storage.QueryOptions{},
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"expected error %v, got %v",
			expectedErr,
			err,
		)
	}
}
