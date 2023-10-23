package plugo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Connection is a user-friendly interface to perform http responses
type Connection interface {
	// HTTP hRequest getter
	Request() *http.Request

	// HTTP Response getter
	Response() *Response

	// URL getter for http.Request.URL()
	URL() *url.URL

	// PathParams gets an slice of parameter values if they exists in the url
	PathParams() []string

	// Param gets the value of a parameter if it exists in the url
	Param(key string) (value string, ok bool)

	// HTML sends a new response in HTML format
	HTML(code int, data string) error

	// String sends a new response in plain text format
	String(code int, data string, args ...any) error

	// JSON sends a new response in JSON format automatically decoding the parameter 'data'
	JSON(code int, data any) error

	// JSONBlob sends a new response in JSON format
	JSONBlob(code int, b []byte) error

	// Blob sends a new response in any desired content-type
	Blob(code int, contentType string, b []byte) error
}

type connectionImpl struct {
	response *Response
	request  *http.Request
	pattern  string
	path     string
	params   []string
}

var _ Connection = &connectionImpl{}

func newConnection(w http.ResponseWriter, r *http.Request) *connectionImpl {
	pattern := r.Context().Value(MethodID("pattern")).(string)

	return &connectionImpl{
		response: NewResponse(w),
		request:  r,
		pattern:  pattern,
		path:     cleanPath(r.URL.Path),
		params:   parseParamKeysFromPattern(pattern),
	}
}

func NewConnection(w http.ResponseWriter, r *http.Request) Connection {
	return newConnection(w, r)
}

func (conn *connectionImpl) Request() *http.Request {
	return conn.request
}

func (conn *connectionImpl) Response() *Response {
	return conn.response
}

func (conn *connectionImpl) URL() *url.URL {
	return conn.request.URL
}

func (conn *connectionImpl) PathParams() []string {
	res := make([]string, 0)

	for _, key := range conn.params {
		pattMoves := strings.Split(strings.Split(conn.pattern, ":"+key)[0], "/")
		if strings.HasSuffix(conn.pattern, "/") {
			pattMoves = pattMoves[:len(pattMoves)-1]
		}

		depth := len(pattMoves) - 1

		moves := strings.Split(conn.path, "/")
		if strings.HasSuffix(conn.path, "/") {
			moves = moves[:len(moves)-1]
		}

		if len(moves) >= depth {
			res = append(res, moves[depth])
		}
	}

	return res
}

func (conn *connectionImpl) Param(key string) (value string, ok bool) {
	values := conn.PathParams()

	for i, param := range conn.params {
		if param == key {
			value = values[i]
			ok = true

			return
		}
	}

	return
}

func (conn *connectionImpl) HTML(code int, data string) error {
	return conn.Blob(code, "text/html", []byte(data))
}

func (conn *connectionImpl) String(code int, data string, args ...any) error {
	return conn.Blob(code, "text/plain", []byte(fmt.Sprintf(data, args...)))
}

func (conn *connectionImpl) JSON(code int, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return conn.JSONBlob(code, b)
}

func (conn *connectionImpl) JSONBlob(code int, b []byte) error {
	return conn.Blob(code, "application/json", b)
}

func (c *connectionImpl) Blob(code int, contentType string, b []byte) error {
	c.writeContentType(contentType)
	if code != http.StatusOK {
		c.response.WriteHeader(code)
	}

	_, err := c.response.Write(b)
	return err
}

func (conn *connectionImpl) writeContentType(value string) {
	header := conn.Response().Header()
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", value)
	}
}
