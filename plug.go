package plugo

import (
	"net/http"
)

// Plug represents a handler for an specific route
type Plug struct {
	serve http.HandlerFunc

	pattern string
}

var _ http.Handler = &Plug{}

func NewPlug(pattern string, serve http.HandlerFunc) *Plug {
	return &Plug{serve, pattern}
}

func (p *Plug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.serve(w, r)
}
