package mock

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

type Server struct {
	Server        *httptest.Server
	ContentHeader string
	//AuthHeader    string
	Method string
	Path   string
	Body   string
	// Toda: are these all in use??
}

// func newTestServer(t *testing.T, responseBody string) (ts *Server) {
func NewServer(responseBody string) (ts *Server) {

	ts = &Server{}

	ts.Server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		body, _ := io.ReadAll(request.Body)
		//assert.NoError(t, err)
		// Todo: detect error??

		ts.ContentHeader = request.Header.Get("Content-Type")
		//ts.AuthHeader = request.Header.Get("Authorization")
		ts.Method = request.Method
		ts.Path = request.RequestURI
		ts.Body = string(body)

		fmt.Fprint(writer, responseBody)
	}))

	return
}
