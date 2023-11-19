// Package basicrt implements the Tripper interface, adding a Basic Auth header.
package basicrt

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// BasicRt implements the Tripper interface.
type BasicRt struct {
	next http.RoundTripper
	auth string
}

// New creates a BasicRt with encoded auth.
func New(username, password string) *BasicRt {

	userpass := fmt.Sprintf("%s:%s", username, password)
	encoded := base64.StdEncoding.EncodeToString([]byte(userpass))

	return &BasicRt{
		auth: fmt.Sprintf("Basic %s", encoded),
	}
}

// Wrap sets the next round tripper, thereby wrapping it.
func (rt *BasicRt) Wrap(next http.RoundTripper) {
	rt.next = next
}

// RoundTrip adds a Basic Auth header to requests.
func (rt *BasicRt) RoundTrip(request *http.Request) (response *http.Response, err error) {

	request.Header.Set("Authorization", rt.auth)

	response, err = rt.next.RoundTrip(request)
	return
}
