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
	FtwHeader     string
	//AuthHeader    string
	Method string
	Path   string
	Body   string
}

// func newTestServer(t *testing.T, responseBody string) (ts *Server) {
func NewServer(responseBody string) (ts *Server) {

	ts = &Server{}

	ts.Server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		body, err := io.ReadAll(request.Body)
		if err != nil {
			panic(err)
			// maybe pass in testy's for more graceful?
		}

		ts.ContentHeader = request.Header.Get("Content-Type")
		ts.FtwHeader = request.Header.Get("ForThe")
		//ts.AuthHeader = request.Header.Get("Authorization")
		ts.Method = request.Method
		ts.Path = request.RequestURI
		ts.Body = string(body)

		fmt.Fprint(writer, responseBody)
	}))

	return
}
