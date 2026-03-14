package giant

import (
	"bufio"
	"context"
	"net/http"
	"slices"

	"github.com/pkg/errors"
)

const (
	// MaxLineSize is the maximum line size for streaming responses.
	MaxLineSize = 64 * 1024 // 64KB default, increase if needed
)

// StreamLines sends a GET request and returns a channel of lines.
// Uses giant's transport but without timeout for long-lived connections.
// Channel closes when context is cancelled, stream ends, or error occurs.
// Example:
//
//	lines, err := client.StreamLines(ctx, "/events")
//	if err != nil {
//	    return err
//	}
//	for data := range lines {
//	    var event DockerEvent
//	    json.Unmarshal(data, &event)
//	    handle(event)
//	}
func (giant *Giant) StreamLines(ctx context.Context, path string) (lines <-chan []byte, err error) {

	rq := Request{
		Method: http.MethodGet,
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

	ch := make(chan []byte)
	go giant.streamLines(ctx, response, ch)
	lines = ch

	return
}

// streamLines reads lines from response body and sends to channel.
// Context cancellation is checked between lines; a blocked Scan() may not
// interrupt immediately until the underlying read completes or errors.
func (giant *Giant) streamLines(ctx context.Context, response *http.Response, ch chan<- []byte) {

	defer close(ch)
	defer response.Body.Close()

	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, MaxLineSize), MaxLineSize)

	for scanner.Scan() {
		line := slices.Clone(scanner.Bytes())

		if giant.Logger != nil {
			giant.Logger.Trace(ctx, "stream line received", "length", len(line))
		}

		select {
		case <-ctx.Done():
			return
		case ch <- line:
		}
	}

	err := scanner.Err()
	if err != nil && giant.Logger != nil {
		err = errors.Wrap(err, "reading from stream")
		giant.Logger.Error(ctx, "stream read failed", err)
	}
}
