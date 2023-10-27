package plugo

import (
	"testing"
)

func TestNodeTree(t *testing.T) {
	tree := newNode("/")
	users := tree.insertNode("users")

	t.Run("check sub nodes", func(t *testing.T) {
		if tree.kind != nodeStatic {
			t.Error("node kind parsing error")
		}

		if users.kind != nodeStatic {
			t.Error("node kind parsing error")
		}

		if len(tree.children) != 1 || len(tree.statics) != 1 {
			t.Error("error inserting a new sub node")
		}
	})

	tree.bind(MethodGet, "/", NewPlug(nil))
	users.bind(MethodPost, "/users", NewPlug(nil))

	t.Run("check endpoints", func(t *testing.T) {
		var endp *endpoint

		endp = tree.endpoints.Value(MethodGet)
		if endp == nil {
			t.Error("error binding method")
		}

		endp = users.endpoints.Value(MethodPost)
		if endp == nil {
			t.Error("error binding method")
		}
	})

	avy := users.insertNode("avatar")
	avy.insertNode(":id")

	t.Run("check node params", func(t *testing.T) {
		if avy.params == nil {
			t.Error("error inserting parametric node")
		}
	})
}
