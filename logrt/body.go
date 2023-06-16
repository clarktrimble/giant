package logrt

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// unexported

// requestBody reads and restores the request body

func requestBody(req *http.Request) (body []byte, err error) {

	body, err = read(req.Body)
	if err != nil {
		return
	}

	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return
}

// responseBody reads and restores the response body

func responseBody(res *http.Response) (body []byte, err error) {

	body, err = read(res.Body)
	if err != nil {
		return
	}

	res.Body = io.NopCloser(bytes.NewBuffer(body))
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
