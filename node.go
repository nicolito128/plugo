package plugo

import (
	"regexp"
	"strings"
)

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

	// parent node
	parent *node

	// catch all nodes
	catchAll *node

	// parametric nodes
	params *node

	// slice with all the nodes
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

func (nd *node) bind(mid MethodID, pattern string, handler HandlerFunc) {
	nd.endpoints[mid] = &endpoint{
		handler,
		pattern,
		parseParamKeysFromPattern(pattern),
	}

	nd.isHandler = true
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
		newElement.label = ":"
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

func (nd *node) findRoute(ctx *contextImpl, search string) *node {
	ok := nd.match(search)
	if ok {
		nd.updateRouteContext(ctx)
		return nd
	}

	if len(nd.children) > 0 {
		for _, child := range nd.children {
			ok = child.match(search)
			if ok {
				child.updateRouteContext(ctx)
				return child
			}
		}
	}

	if nd.params != nil {
		nd.params.updateRouteContext(ctx)
		return nd.params
	}

	if nd.catchAll != nil {
		nd.catchAll.updateRouteContext(ctx)
		return nd.catchAll
	}

	return nil
}

func (nd *node) updateRouteContext(ctx *contextImpl) {
	if ctx != nil {
		ctx.depth += 1

		if nd.kind == nodeParam {
			moves := strings.Split(ctx.path, "/")
			if strings.HasSuffix(ctx.path, "/") {
				moves = moves[:len(moves)-1]
			}

			if len(moves) >= ctx.depth {
				value := moves[ctx.depth-1]
				if value != "" {
					ctx.params = append(ctx.params, value)
				}
			}
		}
	}
}

func parseStringToNodeType(s string) nodeType {
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		return nodeRegexp
	}

	if strings.HasPrefix(s, ":") {
		return nodeParam
	}

	if strings.HasSuffix(s, "*") {
		return nodeCatchAll
	}

	return nodeStatic
}

func parseParamKeysFromPattern(pattern string) []string {
	paramsMatcher := regexp.MustCompile(`/(:[a-zA-Z])\w+/g`)
	maxLen := len(strings.Split(pattern, "/")[1:])

	// Removing ":" from keys
	result := paramsMatcher.FindAllString(pattern, maxLen)
	for i := range result {
		result[i] = strings.Replace(result[i], ":", "", 1)
	}

	return result
}
