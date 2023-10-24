package plugo

import (
	"net/http"
)

// Plug represents a handler for an specific route
type Plug struct {
	serve http.HandlerFunc
}

var _ http.Handler = &Plug{}

func NewPlug(serve http.HandlerFunc) *Plug {
	return &Plug{serve}
}

func (p *Plug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.serve(w, r)
}
