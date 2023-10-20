package plugo

import (
	"bufio"
	"net"
	"net/http"
)

// Response is a wrapper for http.ResponseWriter
type Response struct {
	http.ResponseWriter

	status int
}

func NewResponse(rw http.ResponseWriter) *Response {
	return &Response{rw, 0}
}

func (res *Response) Write(b []byte) (n int, err error) {
	if res.status == 0 {
		res.status = http.StatusOK
	}

	if res.status != http.StatusOK {
		res.WriteHeader(res.status)
	}

	n, err = res.ResponseWriter.Write(b)

	return
}

func (res *Response) WriteHeader(statusCode int) {
	if statusCode != http.StatusOK {
		res.status = statusCode
		res.ResponseWriter.WriteHeader(statusCode)
	}
}

func (res *Response) Status() int {
	return res.status
}

func (res *Response) Unwrap() http.ResponseWriter {
	return res.ResponseWriter
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (res *Response) Flush() {
	res.ResponseWriter.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (res *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return res.ResponseWriter.(http.Hijacker).Hijack()
}
