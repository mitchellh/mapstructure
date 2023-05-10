package mapstructure

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type NamespaceKey interface{}
type NamespaceIdx int
type NamespaceFld string

type Namespace struct {
	items []interface{}
}

func NewNamespace() *Namespace {
	return &Namespace{}
}

func (ns *Namespace) AppendNamespace(namespace Namespace) *Namespace {
	ns.items = append(ns.items, namespace.items...)
	return ns
}

func (ns *Namespace) PrependNamespace(namespace Namespace) *Namespace {
	ns.items = append(namespace.items, ns.items...)
	return ns
}

func (ns *Namespace) AppendKey(keys ...interface{}) *Namespace {
	for _, k := range keys {
		ns.items = append(ns.items, NamespaceKey(k))
	}
	return ns
}

func (ns *Namespace) PrependKey(keys ...interface{}) *Namespace {
	ppns := (&Namespace{}).AppendKey(keys...)
	ns.items = append(ppns.items, ns.items...)
	return ns
}

func (ns *Namespace) AppendIdx(idxs ...int) *Namespace {
	for _, i := range idxs {
		ns.items = append(ns.items, NamespaceIdx(i))
	}
	return ns
}

func (ns *Namespace) PrependIdx(idxs ...int) *Namespace {
	ppns := (&Namespace{}).AppendIdx(idxs...)
	ns.items = append(ppns.items, ns.items...)
	return ns
}

func (ns *Namespace) AppendFld(flds ...string) *Namespace {
	for _, f := range flds {
		ns.items = append(ns.items, NamespaceFld(f))
	}
	return ns
}

func (ns *Namespace) PrependFld(flds ...string) *Namespace {
	ppns := (&Namespace{}).AppendFld(flds...)
	ns.items = append(ppns.items, ns.items...)
	return ns
}

func (ns Namespace) Len() int {
	return len(ns.items)
}

// Get() return the i-th namespace item. If i < 0 return the last item.
func (ns Namespace) Get(i int) interface{} {
	if i < 0 {
		i = len(ns.items) - 1
	}
	if i < 0 || i >= len(ns.items) {
		return nil
	}
	return ns.items[i]
}

// GetAsString() as Get() but return the item string representation
func (ns Namespace) GetAsString(i int) string {
	if item := ns.Get(i); item != nil {
		str := ns.string(item)
		return str
	}
	return ""
}

func (ns Namespace) string(item interface{}) string {
	var result string
	switch value := item.(type) {
	case NamespaceFld:
		result = string(value)
	case NamespaceIdx:
		result = fmt.Sprintf("[%d]", int(value))
	case NamespaceKey:
		result = fmt.Sprintf("[%v]", value)
	}
	return result
}

func (ns Namespace) Format(fldSeparator string, idxSeparator string, keySeparator string) string {
	var result, sep string

	if len(ns.items) > 0 {
		result = ns.string(ns.items[0])
	}
	for i := 1; i < len(ns.items); i++ {
		item := ns.items[i]
		switch item.(type) {
		case NamespaceFld:
			sep = fldSeparator
		case NamespaceIdx:
			sep = idxSeparator
		case NamespaceKey:
			sep = keySeparator
		}
		result += sep + ns.string(item)
	}
	return result
}

func (ns Namespace) String() string {
	return ns.Format(".", "", "")
}

func (ns Namespace) Duplicate() *Namespace {
	return &Namespace{items: ns.items[:]}
}

type DecodingError struct {
	namespace Namespace
	header    string
	error     error
}

func NewDecodingErrorFormat(format string, args ...interface{}) *DecodingError {
	return &DecodingError{
		error: fmt.Errorf(format, args...),
	}
}

func NewDecodingErrorWrap(err error) *DecodingError {
	return &DecodingError{error: err}
}

func (dErr *DecodingError) WithHeader(format string, args ...interface{}) *DecodingError {
	dErr.header = fmt.Sprintf(format, args...)
	return dErr
}

// Duplicate() won't duplicate any wrapped error in DecodingError for it doesn't
// know how to do it without loosing the error type (i.e. via errors.New()).
func (dErr DecodingError) Duplicate() *DecodingError {
	return &DecodingError{
		namespace: *dErr.namespace.Duplicate(),
		error:     dErr.error,
	}
}

func (dErr *DecodingError) GetNamespace() Namespace {
	return *dErr.namespace.Duplicate()
}

func (dErr *DecodingError) SetNamespace(namespace Namespace) *DecodingError {
	dErr.namespace = *namespace.Duplicate()
	return dErr
}

func (dErr *DecodingError) PrependNamespace(ns Namespace) *DecodingError {
	dErr.namespace.PrependNamespace(ns)
	return dErr
}

func (dErr *DecodingError) AppendNamespace(ns Namespace) *DecodingError {
	dErr.namespace.AppendNamespace(ns)
	return dErr
}

func (dErr *DecodingError) Error() string {
	if dErr.namespace.Len() > 0 {
		return fmt.Sprintf("while decoding '%s': %s%s", dErr.namespace, dErr.header, dErr.error.Error())
	}
	return dErr.error.Error()
}

func (dErr *DecodingError) Unwrap() error {
	return dErr.error
}

// Error implements the error interface and can represents multiple
// errors that occur in the course of a single decode.
type DecodingErrors struct {
	errors []DecodingError
}

func NewDecodingErrors() *DecodingErrors {
	return &DecodingErrors{}
}

func (e *DecodingErrors) Len() int {
	return len(e.errors)
}

func (e *DecodingErrors) Get(i int) *DecodingError {
	if i >= len(e.errors) {
		return nil
	}
	return &e.errors[i]
}

func (e *DecodingErrors) Append(err error) *DecodingErrors {
	if err == nil ||
		(reflect.TypeOf(err).Kind() == reflect.Ptr &&
			reflect.ValueOf(err).IsNil()) {
		return e
	}
	switch err_ := err.(type) {
	case *DecodingErrors:
		e.errors = append(e.errors, err_.errors...)
	case *DecodingError:
		e.errors = append(e.errors, *err_)
	default:
		e.errors = append(e.errors, *NewDecodingErrorWrap(e))
	}
	return e
}

// Duplicate() duplicate DecodingErrors by duplicating each DecodedError stored
// in it. Please check also DecodedError.Duplicate()
func (e *DecodingErrors) Duplicate() *DecodingErrors {
	e_ := &DecodingErrors{
		errors: make([]DecodingError, len(e.errors)),
	}
	for i, err := range e.errors {
		e_.errors[i] = *err.Duplicate()
	}
	return e_
}

func (dErr *DecodingErrors) PrependNamespace(ns Namespace) *DecodingErrors {
	errors := dErr.errors
	for i, err := range dErr.errors {
		errors[i] = *err.PrependNamespace(ns)
	}
	return dErr
}

func (dErr *DecodingErrors) AppendNamespace(ns Namespace) *DecodingErrors {
	errors := dErr.errors
	for i, err := range dErr.errors {
		errors[i] = *err.AppendNamespace(ns)
	}
	return dErr
}

func (e *DecodingErrors) Error() string {
	points := make([]string, len(e.errors))
	for i, err := range e.errors {
		points[i] = fmt.Sprintf("* %s", &err)
	}
	sort.Strings(points)
	return fmt.Sprintf("%d error(s) decoding:\n\n%s",
		len(e.errors), strings.Join(points, "\n"))
}

// WrappedErrors implements the errwrap.Wrapper interface to make this
// return value more useful with the errwrap and go-multierror libraries.
func (e *DecodingErrors) WrappedErrors() []error {
	if e == nil {
		return nil
	}
	result := make([]error, len(e.errors))
	for i, e := range e.errors {
		result[i] = e.Duplicate()
	}
	return result
}
