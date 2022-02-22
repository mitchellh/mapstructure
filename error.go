package mapstructure

import (
	"fmt"
	"sort"
	"strings"
)

// Error implements the error interface and can represents multiple
// errors that occur in the course of a single decode.
type Error struct {
	// Deprecated: left for backward compatibility.
	Errors     []string
	realErrors []error
}

func newMultiError(errors []error) *Error {
	stringErrors := make([]string, len(errors))
	for i, err := range errors {
		stringErrors[i] = err.Error()
	}
	return &Error{
		Errors:     stringErrors,
		realErrors: errors,
	}
}

func (e *Error) Error() string {
	points := make([]string, len(e.realErrors))
	for i, err := range e.realErrors {
		points[i] = fmt.Sprintf("* %s", err.Error())
	}

	sort.Strings(points)
	return fmt.Sprintf(
		"%d error(s) decoding:\n\n%s",
		len(e.realErrors), strings.Join(points, "\n"))
}

// WrappedErrors implements the errwrap.Wrapper interface to make this
// return value more useful with the errwrap and go-multierror libraries.
func (e *Error) WrappedErrors() []error {
	if e == nil {
		return nil
	}

	return e.realErrors
}

func appendErrors(errors []string, err error) []string {
	switch e := err.(type) {
	case *Error:
		return append(errors, e.Errors...)
	default:
		return append(errors, e.Error())
	}
}
