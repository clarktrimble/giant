package mock

import (
	"io"
	"net/http"
	"strings"
)

type TestRt struct {
	Status int
}

func (rt *TestRt) RoundTrip(request *http.Request) (response *http.Response, err error) {

	response = &http.Response{
		StatusCode: rt.Status,
		Body:       io.NopCloser(strings.NewReader(`{"ima": "pc"}`)),
		Request:    request,
	}

	return
}
