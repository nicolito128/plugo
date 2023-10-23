package plugo

import (
	"net/http"
)

// Plug represents a handler for an specific route
type Plug struct {
	pattern string
	serve   http.HandlerFunc
}

var _ Plugger = &Plug{}

type Plugger interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	Pattern() string
	Unwrap() *Plug
}

func NewPlug(pattern string, serve http.HandlerFunc) *Plug {
	return &Plug{pattern, serve}
}

func (p *Plug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.serve(w, r)
}

func (p *Plug) Pattern() string {
	return p.pattern
}

func (p *Plug) Unwrap() *Plug {
	return p
}
