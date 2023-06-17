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
	Method        string
	Path          string
	Body          string
}

func NewServer(responseBody string) (ts *Server) {

	ts = &Server{}

	ts.Server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		body, err := io.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}

		ts.ContentHeader = request.Header.Get("Content-Type")
		ts.Method = request.Method
		ts.Path = request.RequestURI
		ts.Body = string(body)

		fmt.Fprint(writer, responseBody)
	}))

	return
}
