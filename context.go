package plugo

import "net/http"

// Context of the http request.
type Context interface {
	// URLParams gets the value of all the parameters in the URL.
	URLParams() []string
	// Param searches for a specific parameter value.
	Param(key string) (value string, ok bool)
}

type contextImpl struct {
	// http request to handle
	req *http.Request

	// current path of the request
	path string

	// how many nodes had to be traversed
	depth int

	// parameter keys
	keyParams []string

	// parameter values
	valueParams []string
}

var _ Context = &contextImpl{}

func newContext(r *http.Request) *contextImpl {
	return &contextImpl{
		req:         r,
		path:        cleanPath(r.URL.Path),
		keyParams:   make([]string, 0),
		valueParams: make([]string, 0),
	}
}

func NewContext(r *http.Request) Context {
	return newContext(r)
}

// URLParams gets the value of all the parameters in the URL.
func (ctx *contextImpl) URLParams() []string {
	return ctx.valueParams
}

// Param searches for a specific parameter value.
func (ctx *contextImpl) Param(key string) (value string, ok bool) {
	keys := ctx.keyParams
	values := ctx.valueParams
	if len(keys) != len(values) || len(keys) == 0 || len(values) == 0 {
		return value, ok
	}

	m := make(map[string]string)
	for i, k := range keys {
		m[k] = values[i]
	}

	value, ok = m[key]
	return value, ok
}
