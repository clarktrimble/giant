package giant

import (
	"context"
	"net/http"
)

type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...any) context.Context
}

type Tripper interface {
	http.RoundTripper
	Wrap(next http.RoundTripper)
}
