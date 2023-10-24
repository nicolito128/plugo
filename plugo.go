package plugo

import (
	"context"
	"net/http"
	"strings"
)

// RouterOption represents a handler for setting Plug configurable parameters.
type RouterOption func(*RouterConfig)

// Router represents an HTTP router.
type Router struct {
	// node tree
	routes *node

	// static nodes
	namedRoutes map[string]*node

	// slice of middlewares to execute before a request
	beforeMiddlewares []MiddlewareFunc

	// slice of middlewares to execute after a request
	afterMiddlewares []MiddlewareFunc

	// public fields to configurate
	*RouterConfig
}

// New creates a new Router pointer with a default set of options
func New(opts ...RouterOption) *Router {
	return NewRouter(opts...)
}

// NewRouter creates a new Router pointer with a default set of options
func NewRouter(opts ...RouterOption) *Router {
	config := &RouterConfig{}
	DefaultRouterOptions(config)
	for _, opt := range opts {
		opt(config)
	}

	router := &Router{
		routes:            newNode(config.IndexPath),
		namedRoutes:       make(map[string]*node),
		beforeMiddlewares: make([]MiddlewareFunc, 0),
		afterMiddlewares:  make([]MiddlewareFunc, 0),
		RouterConfig:      config,
	}

	return router
}

// RouterConfig is a set of public fields to configurate a Router
type RouterConfig struct {
	//
	IndexPath string

	// strick check for '/' at the end of a route
	SlashStrictly bool

	// 404 not found handler
	NotFound http.Handler

	// 405 method not allowed handler
	MethodNotAllowed http.Handler
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ensure execution at the end
	defer func() {
		rt.handleMiddlewares(w, r, rt.afterMiddlewares...)
	}()

	// handling before middlewares
	rt.handleMiddlewares(w, r, rt.beforeMiddlewares...)

	// handling the current request
	route, endp, handler := rt.handleRequest(r)
	if route != nil {
		rt.handleMiddlewares(w, r, route.middlewares...)

		if endp != nil {
			r = r.WithContext(context.WithValue(context.Background(), MethodID("pattern"), endp.pattern))
		}
	}

	handler.ServeHTTP(w, r)
}

// Pre adds a set of middlewares to be executed before a request.
func (rt *Router) Pre(middlewares ...MiddlewareFunc) {
	rt.beforeMiddlewares = append(rt.beforeMiddlewares, middlewares...)
}

// Use adds a set of middlewares to be executed after a request.
func (rt *Router) Use(middlewares ...MiddlewareFunc) {
	rt.afterMiddlewares = append(rt.beforeMiddlewares, middlewares...)
}

// Get registers a new HTTP GET method handler.
func (rt *Router) Get(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	rt.HandleFunc(MethodGet, pattern, handler, middlewares...)
}

// Post registers a new HTTP POST method handler.
func (rt *Router) Post(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	rt.HandleFunc(MethodPost, pattern, handler, middlewares...)
}

// Put registers a new HTTP PUT method handler.
func (rt *Router) Put(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	rt.HandleFunc(MethodPut, pattern, handler, middlewares...)
}

// Delete registers a new HTTP DELETE method handler.
func (rt *Router) Delete(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	rt.HandleFunc(MethodDelete, pattern, handler, middlewares...)
}

// Connect registers a new HTTP CONNECT method handler.
func (rt *Router) Connect(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	rt.HandleFunc(MethodConnect, pattern, handler, middlewares...)
}

// Head registers a new HTTP HEAD method handler.
func (rt *Router) Head(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	rt.HandleFunc(MethodHead, pattern, handler, middlewares...)
}

// Options registers a new HTTP OPTIONS method handler.
func (rt *Router) Options(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	rt.HandleFunc(MethodOptions, pattern, handler, middlewares...)
}

// Trace registers a new HTTP TRACE method handler.
func (rt *Router) Trace(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	rt.HandleFunc(MethodTrace, pattern, handler, middlewares...)
}

// Handle registers a new handler to serve http requests in the provided method.
func (rt *Router) Handle(method MethodID, pattern string, handler http.Handler, middlewares ...MiddlewareFunc) {
	if !method.Allowed() {
		panic(ErrMethodNotAllowed)
	}

	if strings.HasSuffix(rt.IndexPath, "/") {
		pattern = rt.IndexPath[0:len(rt.IndexPath)-1] + pattern
	} else {
		pattern = rt.IndexPath + pattern
	}

	var isStatic bool = true

	// slice of elements splited according to whether slash strictly is true or false
	cleaned := cleanPath(pattern)
	moves := rt.parsePatternToMovements(cleaned)
	// initial node of the tree
	root := rt.routes
	for _, move := range moves {
		if parseStringToNodeType(move) != nodeStatic {
			isStatic = false
		}

		ok := root.match(move)
		if ok {
			continue
		}

		aux := root.findRoute(move)
		// if it exists, then take it as the current root
		if aux != nil {
			root = aux
		} else {
			// if it does not exist, then create it in the current root
			root = root.attach(move)
		}
	}

	root.bind(method, pattern, NewPlug(pattern, handler.ServeHTTP))
	root.use(middlewares...)

	if isStatic {
		rt.namedRoutes[cleaned] = root
	}
}

// HandleFunc registers a new handler function to serve http requests in the provided method.
func (rt *Router) HandleFunc(method MethodID, pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	rt.Handle(method, pattern, NewPlug(pattern, handler))
}

func (rt *Router) handleRequest(r *http.Request) (*node, *endpoint, http.Handler) {
	route, staticOk := rt.namedRoutes[cleanPath(r.URL.Path)]
	if staticOk {
		ent := route.endpoints.Value(MethodID(r.Method))
		if ent == nil {
			return nil, nil, rt.MethodNotAllowed
		}

		return route, ent, ent.handler
	}

	// steps to search a determinate path
	moves := rt.parsePatternToMovements(cleanPath(r.URL.Path))
	// search for node
	root := rt.routes
	for _, move := range moves {
		if root != nil {
			root = root.findRoute(move)
		}
	}

	if root != nil {
		ent := root.endpoints.Value(MethodID(r.Method))
		if ent != nil {
			return root, ent, ent.handler
		}

		if root.catchAll != nil {
			ent = root.catchAll.endpoints.Value(MethodID(r.Method))
			if ent != nil {
				return root, ent, ent.handler
			} else {
				return nil, nil, rt.MethodNotAllowed
			}
		}
	}

	return nil, nil, rt.NotFound
}

func (rt *Router) handleMiddlewares(w http.ResponseWriter, r *http.Request, middlewares ...MiddlewareFunc) {
	for i := range middlewares {
		handler := middlewares[i]()

		handler(w, r)
	}
}

// parsePatternToMovements splits a pattern depending on whether slashStrictly is true or false.
func (rt *Router) parsePatternToMovements(pattern string) []string {
	var moves = make([]string, 0)

	if rt.SlashStrictly {
		moves = strings.SplitAfter(pattern, "/")
	} else {
		moves = strings.Split(pattern, "/")
		moves[0] = "/"
	}

	if strings.HasSuffix(pattern, "/") {
		moves = moves[:len(moves)-1]
	}

	return moves
}

// DefaultPlugOptions sets an basic configuration for the new Plug object.
func DefaultRouterOptions(rt *RouterConfig) {
	rt.IndexPath = "/"

	rt.SlashStrictly = false

	rt.NotFound = &notFoundPlug{}

	rt.MethodNotAllowed = &methodNotAllowedPlug{}
}
