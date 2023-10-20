package plugo

// HandlerFunc type to handle http request
type HandlerFunc func(conn Connection) error

// endpoints is a mapping of http method constants to handlers
type endpoints map[MethodID]*endpoint

type endpoint struct {
	handler   HandlerFunc
	pattern   string
	paramKeys []string
}

func (e endpoints) Value(method MethodID) *endpoint {
	mh, ok := e[method]
	if !ok {
		return nil
	}

	return mh
}
