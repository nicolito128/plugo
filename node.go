package plugo

import (
	"net/http"
	"regexp"
)

// MidlewareFunc represents a function that is executed before or after an http request.
type MiddlewareFunc func(*error) http.HandlerFunc

// node represents a single route node of the tree.
type node struct {
	// type of the node
	kind nodeType

	// pattern expression of the current node
	label string

	// regexp matcher for regexp nodes
	rex *regexp.Regexp

	// http handler endpoints
	endpoints endpoints

	// slice of middlewares to execute after a request
	middlewares []MiddlewareFunc

	// parent node
	parent *node

	// catch all nodes
	catchAll *node

	// parametric nodes
	params *node

	// slice with static and regexp nodes
	children []*node

	// slice of regexp node
	matchers []*node

	// slice of static node
	statics []*node

	// isLeaf indicates that node does not have child routes
	isLeaf bool

	// isHandler indicate if the node has any http handler func registered
	isHandler bool
}

type nodeType uint8

const (
	nodeStatic   nodeType = iota // Ex. /home
	nodeRegexp                   // Ex. /{[0-9]+}
	nodeParam                    // Ex. /:user
	nodeCatchAll                 // Ex. /api/*
)

func newNode(pattern string) *node {
	if len(pattern) == 0 {
		panic(ErrCreateEmptyNode)
	}

	var exp *regexp.Regexp
	var err error

	typ := parseStringToNodeType(pattern)
	if typ == nodeRegexp {
		if len(pattern) >= 2 {
			exp, err = regexp.Compile(pattern[1 : len(pattern)-1])
			if err != nil {
				panic(ErrPatternNotCompile)
			}
		}
	}

	return &node{
		kind:      typ,
		label:     pattern,
		rex:       exp,
		endpoints: make(endpoints),
		catchAll:  nil,
		params:    nil,
		children:  make([]*node, 0),
		matchers:  make([]*node, 0),
		statics:   make([]*node, 0),
	}
}

func (nd *node) bind(mid MethodID, pattern string, handler http.Handler) {
	nd.endpoints[mid] = &endpoint{
		handler,
		pattern,
	}

	nd.isHandler = true
}

func (nd *node) use(middlewares ...MiddlewareFunc) {
	nd.middlewares = append(nd.middlewares, middlewares...)
}

func (nd *node) attach(label string) *node {
	newElement := newNode(label)
	newElement.isLeaf = true
	newElement.parent = nd
	nd.isLeaf = false

	switch newElement.kind {
	case nodeRegexp:
		nd.matchers = append(nd.matchers, newElement)
		nd.children = append(nd.children, newElement)

	case nodeStatic:
		nd.statics = append(nd.statics, newElement)
		nd.children = append(nd.children, newElement)

	case nodeParam:
		// clear for generic parameter
		newElement.label = ":"

		// saves the data of the previous node before overwriting it
		if nd.params != nil {
			newElement.children = append(newElement.children, nd.children...)
			newElement.matchers = append(newElement.matchers, nd.matchers...)
			newElement.statics = append(newElement.statics, nd.statics...)
			newElement.catchAll = nd.catchAll
			newElement.endpoints = nd.endpoints
		}

		nd.params = newElement

	case nodeCatchAll:
		nd.catchAll = newElement
	}

	return newElement
}

// match checks for matches on static and regexp nodes.
func (nd *node) match(exp string) bool {
	switch nd.kind {
	case nodeStatic:
		if nd.label == exp {
			return true
		}

	case nodeRegexp:
		if nd.rex != nil {
			ok := nd.rex.MatchString(exp)
			if ok {
				return true
			}
		}
	}

	return false
}

func (nd *node) findRoute(search string) *node {
	ok := nd.match(search)
	if ok {
		return nd
	}

	if len(nd.children) > 0 {
		for _, child := range nd.children {
			ok = child.match(search)
			if ok {
				return child
			}
		}
	}

	if nd.params != nil {
		return nd.params
	}

	if nd.catchAll != nil {
		return nd.catchAll
	}

	return nil
}
