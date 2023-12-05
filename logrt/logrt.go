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

// Todo: giant.Logger ifc is awkward? prolly dont want/need extra pkg's here?
type Logger interface {
	Info(ctx context.Context, msg string, kv ...any)
	Error(ctx context.Context, msg string, err error, kv ...any)
	WithFields(ctx context.Context, kv ...any) context.Context
}

const (
	idLen int = 7
)

var (
	RedactHeaders = map[string]bool{
		"Authorization": true,
	}
)

// LogRt implements the Tripper interface
type LogRt struct {
	Logger Logger
	next   http.RoundTripper
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

	rt.Logger.Info(ctx, "sending request", requestFields(request)...)

	response, err = rt.next.RoundTrip(request)
	if err != nil {
		return
	}

	rt.Logger.Info(ctx, "received response", responseFields(response, request.URL.Path, start)...)

	return
}

var SkipBody bool

// unexported

func requestFields(request *http.Request) (fields []any) {

	fields = []any{
		"method", request.Method,
		"scheme", request.URL.Scheme,
		"host", request.URL.Host,
		"path", request.URL.Path,
		"headers", redact(request.Header),
		"query", request.URL.Query(),
	}

	if !SkipBody {

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

func responseFields(response *http.Response, path string, start time.Time) (fields []any) {

	fields = []any{
		"status", response.StatusCode,
		"path", path,
		"headers", response.Header,
		"elapsed", time.Since(start),
	}

	if !SkipBody {

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

func redact(header http.Header) (redacted http.Header) {

	redacted = header.Clone()
	for key := range header {

		redacted[key] = header[key]
		if RedactHeaders[key] {
			redacted[key] = []string{"--redacted--"}
		}
	}

	return
}
