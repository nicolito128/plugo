package plugo

import "net/http"

// Context of the http request.
type Context interface {
	URLParams() []string
}

type contextImpl struct {
	//
	req *http.Request

	//
	path string

	//
	depth int

	//
	params []string
}

var _ Context = &contextImpl{}

func newContext(r *http.Request) *contextImpl {
	return &contextImpl{
		req:    r,
		path:   cleanPath(r.URL.Path),
		params: make([]string, 0),
	}
}

func NewContext(r *http.Request) Context {
	return newContext(r)
}

func (ctx *contextImpl) URLParams() []string {
	return ctx.params
}
