package plugo

import (
	"net/http"
)

// HTTP Method Wrapper
type MethodID string

func (mid MethodID) Match(exp string) bool {
	return string(mid) == exp
}

func (mid MethodID) Allowed() bool {
	switch mid {
	case
		MethodGet,
		MethodPost,
		MethodPut,
		MethodDelete,
		MethodConnect,
		MethodHead,
		MethodOptions,
		MethodTrace:

		return true
	}

	return false
}

// net/http method wrapped
const (
	MethodGet     MethodID = http.MethodGet
	MethodPost             = http.MethodPost
	MethodPut              = http.MethodPut
	MethodDelete           = http.MethodDelete
	MethodConnect          = http.MethodConnect
	MethodHead             = http.MethodHead
	MethodOptions          = http.MethodOptions
	MethodTrace            = http.MethodTrace
)
