package plugo

import (
	"errors"
)

var ErrMethodNotAllowed = errors.New("method not allowed for http request")

var ErrHandleEmptyPattern = errors.New("could not handle an empty pattern")

var ErrCreateEmptyNode = errors.New("could not create a new node with an empty pattern")

var ErrPatternNotCompile = errors.New("could not compile pattern to a valid regular expression")
