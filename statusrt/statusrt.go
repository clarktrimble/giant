// Package statusrt implements the Tripper interface
// returning an error if the reponse status code is not in the 200's
package statusrt

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
)

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
func (rt *StatusRt) RoundTrip(request *http.Request) (response *http.Response, err error) {

	response, err = rt.next.RoundTrip(request)
	if err != nil {
		return
	}

	if !validStatusCode(response.StatusCode) {
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			err = errors.Wrapf(readErr, "somehow failed to read body after unexpected status code %d", response.StatusCode)
		} else {
			err = errors.Errorf("unexpected status code %d with body: %s", response.StatusCode, body)
		}
	}
	return
}

//
// unexported
//

func validStatusCode(statusCode int) bool {

	// suppport more variation as needed
	return statusCode >= minStatus && statusCode < maxStatus
}
