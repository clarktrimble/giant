// Package logrt holds space for an implementation of http.RoundTripper that logs requests and responses.
package logrt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/clarktrimble/hondo"
	"github.com/pkg/errors"
)

const (
	idLen int = 7
)

// Logger specifies a logger.
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...any) context.Context
}

// LogRt implements the Tripper interface logging requests and responses.
type LogRt struct {
	RedactHeaders map[string]bool
	SkipBody      bool
	Logger        Logger
	next          http.RoundTripper
}

// New creates a LogRt.
func New(lgr Logger, redactHeaders []string, skipBody bool) (logRt *LogRt) {

	logRt = &LogRt{
		RedactHeaders: map[string]bool{},
		SkipBody:      skipBody,
		Logger:        lgr,
	}

	// always redact for basic auth

	for _, key := range append(redactHeaders, "Authorization") {
		logRt.RedactHeaders[key] = true
	}

	return
}

// Wrap sets the next round tripper, thereby wrapping it
func (rt *LogRt) Wrap(next http.RoundTripper) {
	rt.next = next
}

// RoundTrip logs the request and response
func (rt *LogRt) RoundTrip(request *http.Request) (response *http.Response, err error) {

	start := time.Now()

	ctx := request.Context()
	ctx = rt.Logger.WithFields(ctx, "request_id", hondo.Rand(idLen))
	request = request.WithContext(ctx)

	// Todo: passthru

	rt.Logger.Info(ctx, "sending request", rt.requestFields(request)...)

	response, err = rt.next.RoundTrip(request)
	if err != nil {
		return
	}

	rt.Logger.Info(ctx, "received response", rt.responseFields(response, start)...)

	return
}

// unexported

func (rt *LogRt) requestFields(request *http.Request) (fields []any) {

	fields = []any{
		"method", request.Method,
		"scheme", request.URL.Scheme,
		"host", request.URL.Host,
		"path", request.URL.Path,
		"headers", rt.redact(request.Header),
		"query", request.URL.Query(),
	}

	if !rt.SkipBody {

		// read body and put it back

		body, err := read(request.Body)
		if err != nil {
			body = []byte(fmt.Sprintf("error: %s", err))
		}
		request.Body = io.NopCloser(bytes.NewBuffer(body))

		fields = append(fields, "body")
		fields = append(fields, string(body))
	}

	return
}

// func (rt *LogRt) responseFields(response *http.Response, path string, start time.Time) (fields []any) {
func (rt *LogRt) responseFields(response *http.Response, start time.Time) (fields []any) {

	fields = []any{
		"status", response.StatusCode,
		"headers", response.Header,
		"elapsed", time.Since(start),
	}

	if response.Request != nil && response.Request.URL != nil {
		fields = append(fields, "path")
		fields = append(fields, response.Request.URL.Path)
	}

	if !rt.SkipBody {

		// read body and put it back

		body, err := read(response.Body)
		if err != nil {
			body = []byte(fmt.Sprintf("error: %s", err))
		}
		response.Body = io.NopCloser(bytes.NewBuffer(body))

		fields = append(fields, "body")
		fields = append(fields, string(body))
	}

	return
}

// read reads ;|
// returning nil if nothing read

func read(reader io.Reader) (data []byte, err error) {

	if reader == nil {
		return
	}

	data, err = io.ReadAll(reader)
	if err != nil {
		err = errors.Wrapf(err, "failed to read from: %#v", reader)
		return
	}
	if len(data) == 0 {
		data = nil
	}

	return
}

func (rt *LogRt) redact(header http.Header) (redacted http.Header) {

	redacted = header.Clone()
	for key := range header {

		redacted[key] = header[key]
		if rt.RedactHeaders[key] {
			redacted[key] = []string{"--redacted--"}
		}
	}

	return
}
