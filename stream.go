package giant

import (
	"bufio"
	"context"
	"iter"
	"net/http"
	"slices"

	"github.com/pkg/errors"
)

// Todo: consider context cancellation handling - check ctx.Err() in All() loop
// or verify http.NewRequestWithContext already handles it

// Lines represents a streaming response where each line is a separate item.
// Common for NDJSON streams, log tailing, or other line-delimited protocols.
// Example:
//
//	lines, err := client.StreamLines(ctx, "GET", "/events")
//	if err != nil {
//	    return err
//	}
//	defer lines.Close()
//
//	for data, err := range lines.All() {
//	    if err != nil {
//	        return err
//	    }
//	    var event DockerEvent
//	    json.Unmarshal(data, &event)
//	    handle(event)
//	}
type Lines struct {
	response *http.Response
	scanner  *bufio.Scanner
}

// StreamLines sends a request and returns Lines for iterating over lines.
// Uses giant's transport but without timeout for long-lived connections.
// Caller must call Close when done.
func (giant *Giant) StreamLines(ctx context.Context, method, path string) (lines *Lines, err error) {

	rq := Request{
		Method: method,
		Path:   path,
	}

	request, err := rq.httpRequest(ctx, giant.BaseUri)
	if err != nil {
		err = errors.Wrapf(err, "failed to create stream request to %s%s", giant.BaseUri, path)
		return
	}

	for key, val := range giant.Headers {
		request.Header.Set(key, val)
	}

	client := &http.Client{
		Transport: giant.Client.Transport, // nil falls back to default in stdlib
	}

	response, err := client.Do(request)
	if err != nil {
		if response != nil {
			response.Body.Close()
		}
		err = errors.Wrapf(err, "stream request to %s%s failed", giant.BaseUri, path)
		return
	}

	lines = &Lines{
		response: response,
		scanner:  bufio.NewScanner(response.Body),
	}
	return
}

// SetBuffer sets the maximum line size. Must be called before All().
// Default is 64KB.
func (lines *Lines) SetBuffer(size int) {

	lines.scanner.Buffer(make([]byte, size), size)
}

// All returns an iterator over lines in the stream as raw bytes.
func (lines *Lines) All() iter.Seq2[[]byte, error] {

	return func(yield func([]byte, error) bool) {
		for lines.scanner.Scan() {
			if !yield(slices.Clone(lines.scanner.Bytes()), nil) {
				return
			}
		}
		if err := lines.scanner.Err(); err != nil {
			yield(nil, err)
		}
	}
}

// Close closes the underlying response body.
func (lines *Lines) Close() error {

	return lines.response.Body.Close()
}
