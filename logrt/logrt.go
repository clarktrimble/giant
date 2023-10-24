// Package logrt holds space for an implementation of http.RoundTripper that logs requests and responses.
package logrt

import (
	"net/http"
	"time"

	"github.com/clarktrimble/giant"
	"github.com/clarktrimble/hondo"
)

// Todo: giant.Logger ifc is awkward? prolly dont want/need extra pkg's here?

const (
	idLen int = 7
)

var (
	RedactHeaders = map[string]bool{}
)

// LogRt implements the Tripper interface
type LogRt struct {
	Logger giant.Logger
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

	reqBody, err := requestBody(request)
	if err != nil {
		rt.Logger.Error(ctx, "roundtrip logger failed to get request body", err)
	}

	// Todo: passthru
	// Todo: redact from headers

	rt.Logger.Info(ctx, "sending request",
		"method", request.Method,
		"scheme", request.URL.Scheme,
		"host", request.URL.Host,
		"path", request.URL.Path,
		"headers", redact(request.Header),
		"query", request.URL.Query(),
		"body", string(reqBody),
	)

	response, err = rt.next.RoundTrip(request)
	if err != nil {
		return
	}

	resBody, err := responseBody(response)
	if err != nil {
		rt.Logger.Error(ctx, "roundtrip logger failed to get response body", err)
	}

	rt.Logger.Info(ctx, "received response",
		"status", response.StatusCode,
		"path", request.URL.Path,
		"headers", response.Header,
		"body", string(resBody),
		"elapsed", time.Since(start),
	)

	return
}

// unexported

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
