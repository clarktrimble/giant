// Package statusrt implements the Tripper interface
// returning an error if the response status code is not in the 200's
package statusrt

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// caveat, from RoundTripper doc:

// RoundTrip should not attempt to interpret the response. In
// particular, RoundTrip must return err == nil if it obtained
// a response, regardless of the response's HTTP status code.

// which is exactly what's being done here!!
// "seems" to work ok .. too bad it doesn't say why
// the idea is to get an error which can be logged on non-200 and give up quick
// when we know non-200 is bust

const (
	minStatus int = 200
	maxStatus int = 300
)

// StatusRt implements RoundTripper.
type StatusRt struct {
	next http.RoundTripper
}

// Wrap sets the next round tripper, thereby wrapping it
func (rt *StatusRt) Wrap(next http.RoundTripper) {
	rt.next = next
}

// RoundTrip returns an error if the status code is bad.
func (rt *StatusRt) RoundTrip(request *http.Request) (*http.Response, error) {

	response, err := rt.next.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if !validStatusCode(response.StatusCode) {
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			err = errors.Wrapf(readErr, "somehow failed to read body after unexpected status code %d", response.StatusCode)
		} else {
			err = errors.Errorf("unexpected status code %d with body: %s", response.StatusCode, body)
		}
		response.Body.Close()
		return nil, err
	}
	return response, nil
}

//
// unexported
//

func validStatusCode(statusCode int) bool {

	// suppport more variation as needed
	return statusCode >= minStatus && statusCode < maxStatus
}
