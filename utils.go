package plugo

import (
	"net/http"
	"path"
	"regexp"
	"strings"
)

// Default plug for MethodNotAllowed http response
type methodNotAllowedPlug struct {
	*Plug
}

func (mna *methodNotAllowedPlug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(405)
	w.Write([]byte("Method not allowed."))
}

// Default plug for NotFound http response
type notFoundPlug struct {
	*Plug
}

func (nf *notFoundPlug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
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
	paramsMatcher := regexp.MustCompile(`(:[a-zA-Z0-9])\w+`)
	moves := strings.Split(pattern, "/")
	if strings.HasSuffix(pattern, "/") {
		moves = moves[1:]
	}

	maxLen := len(moves)

	// Removing ":" from keys
	result := paramsMatcher.FindAllString(pattern, maxLen)
	for i := range result {
		result[i] = strings.Replace(result[i], ":", "", 1)
	}

	return result
}
