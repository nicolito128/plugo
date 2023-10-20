package plugo

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// Connection is a user-friendly interface to perform http responses
type Connection interface {
	// Request context utilities
	Context() Context

	// HTTP Request getter
	Request() *http.Request

	// HTTP Response getter
	Response() *Response

	// Getter for http.Request.URL
	URL() *url.URL

	// Call to Connection#Context().URLParams method
	URLParams() []string

	// Call to Connection#Context().Param method
	Param(key string) (value string, ok bool)

	// Send a new response in HTML format
	HTML(code int, data string) error

	// Send a new response in plain text format
	String(code int, data string) error

	// Send a new response in JSON format automatically decoding the parameter 'data'
	JSON(code int, data any) error

	// Send a new response in JSON format
	JSONBlob(code int, b []byte) error

	// Send a new response in any desired Content-Type
	Blob(code int, contentType string, b []byte) error
}

type connectionImpl struct {
	ctx      Context
	response *Response
	request  *http.Request
}

var _ Connection = &connectionImpl{}

func NewConnection(ctx Context, w http.ResponseWriter, req *http.Request) *connectionImpl {
	return &connectionImpl{ctx, NewResponse(w), req}
}

func (conn *connectionImpl) Context() Context {
	return conn.ctx
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

func (conn *connectionImpl) URLParams() []string {
	return conn.ctx.URLParams()
}

func (conn *connectionImpl) Param(key string) (value string, ok bool) {
	return conn.ctx.Param(key)
}

func (conn *connectionImpl) HTML(code int, data string) error {
	return conn.Blob(code, "text/html", []byte(data))
}

func (conn *connectionImpl) String(code int, data string) error {
	return conn.Blob(code, "text/plain", []byte(data))
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

func (c *connectionImpl) Blob(code int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	if code != http.StatusOK {
		c.response.WriteHeader(code)
	}

	_, err = c.response.Write(b)

	return
}

func (conn *connectionImpl) writeContentType(value string) {
	header := conn.Response().Header()
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", value)
	}
}
