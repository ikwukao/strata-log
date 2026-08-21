package ingest

import "errors"

var (
	// ErrProcessorClosed indicates that the ingestion processor is shutting down
	// and cannot accept additional log entries.
	ErrProcessorClosed = errors.New("processor is closed")

	// ErrProcessorFull indicates that the processor cannot accept more work
	// because its configured capacity has been exhausted.
	ErrProcessorFull = errors.New("processor is full")
)

// Processor accepts validated log entries for asynchronous processing.
type Processor interface {
	Submit(LogEntry) error
}
