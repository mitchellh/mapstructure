package mapstructure

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// FieldError implements the error interface and provide access to
// field path where that error occurred.
type FieldError interface {
	error
	Path() FieldPath
}

// baseError default implementation of FieldError.
type baseError struct {
	error
	path FieldPath
}

func (e baseError) Path() FieldPath {
	return e.path
}

func formatError(path FieldPath, format string, a ...interface{}) error {
	return &baseError{
		path:  path,
		error: fmt.Errorf(format, a...),
	}
}

// SliceExpectedError implements the error interface. It provides access to
// field path where that error occurred and specific for this type of errors parameters.
type SliceExpectedError struct {
	path FieldPath
	Got  reflect.Kind
}

func (e SliceExpectedError) Path() FieldPath {
	return e.path
}

func (e *SliceExpectedError) Error() string {
	return fmt.Sprintf(
		"'%s': source data must be an array or slice, got %s", e.path, e.Got)
}

// UnexpectedUnconvertibleTypeError implements the error interface. It provides access to
// field path where that error occurred and specific for this type of errors parameters.
type UnexpectedUnconvertibleTypeError struct {
	path     FieldPath
	Expected reflect.Type
	Got      reflect.Type
	Data     interface{}
}

func (e UnexpectedUnconvertibleTypeError) Path() FieldPath {
	return e.path
}

func (e *UnexpectedUnconvertibleTypeError) Error() string {
	return fmt.Sprintf(
		"'%s' expected type '%s', got unconvertible type '%s', value: '%v'",
		e.path, e.Expected, e.Got, e.Data)
}

// CanNotParseError implements the error interface. It provides access to
// field path where that error occurred and specific for this type of errors parameters.
type CanNotParseError struct {
	path   FieldPath
	Reason error
	Type   string
}

func (e CanNotParseError) Path() FieldPath {
	return e.path
}

func (e *CanNotParseError) Error() string {
	return fmt.Sprintf("cannot parse '%s' as %s: %s", e.path, e.Type, e.Reason)
}

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
