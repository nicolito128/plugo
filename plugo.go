package plugo

import (
	"log"
	"net/http"
	"path"
	"strings"
)

// ErrorHandler represents a handler to catch when an error occurs in an HTTP request.
type ErrorHandler func(Connection, error)

// MidlewareFunc represents a function that is executed before or after an http request.
type MiddlewareFunc func() HandlerFunc

// PlugOption represents a handler for setting Plug configurable parameters.
type PlugOption func(*Plug)

// Plug is an HTTP router.
type Plug struct {
	// http serve mux
	mux *http.ServeMux

	// node tree
	routes *node

	// static nodes
	namedRoutes map[string]*node

	// slice of middlewares to execute before a request
	beforeMiddlewares []MiddlewareFunc

	// slice of middlewares to execute after a request
	afterMiddlewares []MiddlewareFunc

	// strick check for '/' at the end of a route
	SlashStrictly bool

	// error handler for http requests
	CatchError ErrorHandler

	// error handler for internal errors (middlewares)
	InternalError ErrorHandler

	// 404 not found handler
	NotFound HandlerFunc

	// 405 method not allowed handler
	MethodNotAllowed HandlerFunc
}

var _ Router = &Plug{}

// Router is an HTTP request handler
type Router interface {
	// ServeHTTP dispatches the request to the handler whose pattern most closely matches the request URL.
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	// Handle registers a new handler to serve http requests in the provided method.
	Handle(method MethodID, pattern string, handler http.Handler, middlewares ...MiddlewareFunc)

	// HandleFunc registers a new handler function to serve http requests in the provided method.
	HandleFunc(method MethodID, pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc)

	// Prev adds a set of middlewares to be executed before a request.
	Prev(middlewares ...MiddlewareFunc)

	// Use adds a set of middlewares to be executed after a request.
	Use(middlewares ...MiddlewareFunc)

	// Get registers a new HTTP GET method handler.
	Get(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc)

	// Post registers a new HTTP POST method handler.
	Post(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc)

	// Put registers a new HTTP PUT method handler.
	Put(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc)

	// Delete registers a new HTTP DELETE method handler.
	Delete(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc)

	// Connect registers a new HTTP CONNECT method handler.
	Connect(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc)

	// Head registers a new HTTP HEAD method handler.
	Head(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc)

	// Options registers a new HTTP OPTIONS method handler.
	Options(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc)

	// Trace registers a new HTTP TRACE method handler.
	Trace(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc)
}

// New creates a new Plug structure as Router with a default configuration.
func New(opts ...PlugOption) Router {
	return NewPlug(opts...)
}

// NewPlug creates a new Plug with a default configuration.
func NewPlug(opts ...PlugOption) *Plug {
	p := &Plug{}

	// Setting options
	DefaultPlugOptions(p)
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// ServeHTTP dispatches the request to the handler whose pattern most closely matches the request URL.
func (p *Plug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r)
	conn := NewConnection(ctx, w, r)

	var err error

	// ensure execution at the end
	defer func() {
		err = p.handleMiddlewares(conn, p.afterMiddlewares...)
		if err != nil {
			p.InternalError(conn, err)
			return
		}
	}()

	// handling before middlewares
	err = p.handleMiddlewares(conn, p.beforeMiddlewares...)
	if err != nil {
		p.InternalError(conn, err)
		return
	}

	// handling the current request
	route, h := p.handleRequest(ctx, r)

	if route != nil {
		err = p.handleMiddlewares(conn, route.middlewares...)
		if err != nil {
			p.InternalError(conn, err)
			return
		}
	}

	if conn.Closed() {
		return
	}

	err = h(conn)
	if err != nil {
		p.CatchError(conn, err)
		return
	}
}

// Handle registers a new handler to serve http requests in the provided method.
func (p *Plug) Handle(method MethodID, pattern string, handler http.Handler, middlewares ...MiddlewareFunc) {
	p.HandleFunc(method, pattern, func(conn Connection) error {
		handler.ServeHTTP(conn.Response().Unwrap(), conn.Request())
		return nil
	}, middlewares...)
}

// HandleFunc registers a new handler function to serve http requests in the provided method.
func (p *Plug) HandleFunc(method MethodID, pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	if !method.Allowed() {
		panic(ErrMethodNotAllowed)
	}

	var isStatic bool = true

	// slice of elements splited according to whether slash strictly is true or false
	cleaned := cleanPath(pattern)
	moves := p.parsePatternToMovements(cleaned)
	// initial node of the tree
	root := p.routes
	for _, move := range moves {
		if parseStringToNodeType(move) != nodeStatic {
			isStatic = false
		}

		ok := root.match(move)
		if ok {
			continue
		}

		aux := root.findRoute(nil, move)
		// if it exists, then take it as the current root
		if aux != nil {
			root = aux
		} else {
			// if it does not exist, then create it in the current root
			root = root.attach(move)
		}
	}

	root.bind(method, pattern, handler)
	root.use(middlewares...)
	if isStatic {
		p.namedRoutes[cleaned] = root
	}
}

// Prev adds a set of middlewares to be executed before a request.
func (p *Plug) Prev(middlewares ...MiddlewareFunc) {
	p.beforeMiddlewares = append(p.beforeMiddlewares, middlewares...)
}

// Use adds a set of middlewares to be executed after a request.
func (p *Plug) Use(middlewares ...MiddlewareFunc) {
	p.afterMiddlewares = append(p.beforeMiddlewares, middlewares...)
}

// Get registers a new HTTP GET method handler.
func (p *Plug) Get(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	p.HandleFunc(MethodGet, pattern, handler, middlewares...)
}

// Post registers a new HTTP POST method handler.
func (p *Plug) Post(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	p.HandleFunc(MethodPost, pattern, handler, middlewares...)
}

// Put registers a new HTTP PUT method handler.
func (p *Plug) Put(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	p.HandleFunc(MethodPut, pattern, handler, middlewares...)
}

// Delete registers a new HTTP DELETE method handler.
func (p *Plug) Delete(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	p.HandleFunc(MethodDelete, pattern, handler, middlewares...)
}

// Connect registers a new HTTP CONNECT method handler.
func (p *Plug) Connect(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	p.HandleFunc(MethodConnect, pattern, handler, middlewares...)
}

// Head registers a new HTTP HEAD method handler.
func (p *Plug) Head(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	p.HandleFunc(MethodHead, pattern, handler, middlewares...)
}

// Options registers a new HTTP OPTIONS method handler.
func (p *Plug) Options(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	p.HandleFunc(MethodOptions, pattern, handler, middlewares...)
}

// Trace registers a new HTTP TRACE method handler.
func (p *Plug) Trace(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	p.HandleFunc(MethodTrace, pattern, handler, middlewares...)
}

// parsePatternToMovements splits and clears a pattern depending on whether slashStrictly is true or false.
func (p *Plug) parsePatternToMovements(pattern string) []string {
	moves := make([]string, 0)
	if p.SlashStrictly {
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

func (p *Plug) handleRequest(ctx *contextImpl, r *http.Request) (*node, HandlerFunc) {
	route, staticOk := p.namedRoutes[cleanPath(r.URL.Path)]
	if staticOk {
		ent := route.endpoints.Value(MethodID(r.Method))
		if ent == nil {
			return route, p.MethodNotAllowed
		}

		return route, ent.handler
	}

	// steps to search a determinate path
	moves := p.parsePatternToMovements(cleanPath(r.URL.Path))
	// search for node
	root := p.routes
	for _, move := range moves {
		if root != nil {
			root = root.findRoute(ctx, move)
		}
	}

	if root != nil {
		ent := root.endpoints.Value(MethodID(r.Method))
		if ent == nil {
			if root.catchAll != nil {
				ent = root.catchAll.endpoints.Value(MethodID(r.Method))
			}

			if ent == nil {
				return root, p.MethodNotAllowed
			}
		}

		return root, ent.handler
	}

	return nil, p.NotFound
}

func (p *Plug) handleMiddlewares(conn Connection, middlewares ...MiddlewareFunc) error {
	if conn.Closed() {
		return nil
	}

	var err error
	for i := range middlewares {
		handler := middlewares[i]()

		if err = handler(conn); err != nil {
			return err
		}
	}

	return err
}

// DefaultPlugOptions sets an basic configuration for the new Plug object.
func DefaultPlugOptions(p *Plug) {
	p.mux = http.NewServeMux()

	p.routes = newNode("/")

	p.namedRoutes = make(map[string]*node)

	p.SlashStrictly = false

	p.CatchError = defaultCatchErrorHandler

	p.InternalError = defaultInternalErrorHandler

	p.NotFound = defaultNotFoundHandler

	p.MethodNotAllowed = defaultMethodNotAllowedHandler
}

func defaultCatchErrorHandler(conn Connection, err error) {
	if err == nil {
		return
	}

	r := conn.Request()
	log.Printf("ERROR! / Method: %s / Status: %s / URI: %s\n", r.Method, r.Response.Status, r.RequestURI)
	log.Println(err)
}

func defaultInternalErrorHandler(conn Connection, err error) {
	if err == nil {
		return
	}

	conn.String(http.StatusInternalServerError, "Internal server error.")
}

func defaultNotFoundHandler(conn Connection) error {
	w := conn.Response().Unwrap()
	r := conn.Request()
	http.NotFound(w, r)

	return nil
}

func defaultMethodNotAllowedHandler(conn Connection) error {
	w := conn.Response().Unwrap()
	w.WriteHeader(405)
	w.Write([]byte("Method not allowed."))

	return nil
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	if p[0] != '/' {
		p = "/" + p
	}

	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}

	return np
}
