package plugo

import "testing"

func TestCleanPath(t *testing.T) {
	var tests = []struct {
		name string
		path string
		want string
	}{
		{"check empty root", "", "/"},
		{"adding slash at the beginning", "foo/bar", "/foo/bar"},
		{"trailing slash back", "/hello/plugo/", "/hello/plugo/"},
	}

	for _, test := range tests {
		got := cleanPath(test.path)
		if got != test.want {
			t.Errorf("%s got '%s' want '%s'", test.name, got, test.want)
		}
	}
}

func TestParseParamKeysFromPattern(t *testing.T) {
	var tests = []struct {
		name    string
		pattern string
		want    []string
	}{
		{"one param", "/hello/:world", []string{"world"}},
		{"two params sequentially", "/users/:id/:name", []string{"id", "name"}},
		{"more params", "/friends/:id/photos/:folder/:name", []string{"id", "folder", "name"}},
	}

	for _, test := range tests {
		got := parseParamKeysFromPattern(test.pattern)

		if len(got) != len(test.want) {
			t.Errorf("%s got %v want %v", test.name, got, test.want)
		}

		for i := range got {
			key := got[i]
			if test.want[i] != key {
				t.Errorf("%s got key %s want key %s", test.name, key, test.want[i])
			}
		}
	}
}
