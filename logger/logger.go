// Package logger defines the logging interface for giant.
package logger

import "context"

// Logger is the interface consumers must implement for giant logging.
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Trace(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...any) context.Context
}
