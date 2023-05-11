package mapstructure

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type NamespaceKey interface{}
type NamespaceIdx int

type NamespaceFld struct {
	useName bool
	name    string
	tag     string
}

func NewNamespaceFld() *NamespaceFld {
	return &NamespaceFld{useName: true}
}

func (nf *NamespaceFld) UseName(useFieldName bool) *NamespaceFld {
	nf.useName = useFieldName
	return nf
}

func (nf *NamespaceFld) SetName(name string) *NamespaceFld {
	nf.name = name
	return nf
}

func (nf *NamespaceFld) SetTag(tag string) *NamespaceFld {
	nf.tag = tag
	return nf
}

func (nf *NamespaceFld) GetName() string {
	return nf.name
}

func (nf *NamespaceFld) GetTag() string {
	return nf.tag
}

func (nf *NamespaceFld) String() string {
	if nf.useName {
		return nf.name
	}
	return nf.tag
}

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

func (ns *Namespace) AppendFld(flds ...NamespaceFld) *Namespace {
	for _, f := range flds {
		ns.items = append(ns.items, f)
	}
	return ns
}

func (ns *Namespace) PrependFld(flds ...NamespaceFld) *Namespace {
	ppns := (&Namespace{}).AppendFld(flds...)
	ns.items = append(ppns.items, ns.items...)
	return ns
}

func (ns *Namespace) UseFldName(useFldName bool) *Namespace {
	for i, item := range ns.items {
		if fld, ok := item.(NamespaceFld); ok {
			ns.items[i] = *fld.UseName(useFldName)
		}
	}
	return ns
}

func (ns *Namespace) Len() int {
	return len(ns.items)
}

// Get() return the i-th namespace item. If i < 0 return the last item.
func (ns *Namespace) Get(i int) interface{} {
	if i < 0 {
		i = len(ns.items) - 1
	}
	if i < 0 || i >= len(ns.items) {
		return nil
	}
	return ns.items[i]
}

// GetAsString() as Get() but return the item string representation
func (ns *Namespace) GetAsString(i int) string {
	if item := ns.Get(i); item != nil {
		str := ns.string(item)
		return str
	}
	return ""
}

func (ns *Namespace) string(item interface{}) string {
	var result string
	switch value := item.(type) {
	case NamespaceFld:
		result = value.String()
	case NamespaceIdx:
		result = fmt.Sprintf("[%d]", int(value))
	case NamespaceKey:
		result = fmt.Sprintf("[%v]", value)
	}
	return result
}

func (ns *Namespace) Format(fldSeparator string, idxSeparator string, keySeparator string) string {
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

func (ns *Namespace) String() string {
	return ns.Format(".", "", "")
}

func (ns *Namespace) Duplicate() *Namespace {
	return &Namespace{items: ns.items[:]}
}

type LocalizedError interface {
	SetNamespace(ns Namespace) LocalizedError
	PrependNamespace(ns Namespace) LocalizedError
	AppendNamespace(ns Namespace) LocalizedError
	SetNamespaceUseFieldName(useFieldName bool) LocalizedError
	Error() string
}

func AsLocalizedError(err error) LocalizedError {
	if e, ok := err.(LocalizedError); ok {
		return e
	}
	return AsDecodingError(err)
}

type DecodingError struct {
	namespace Namespace // namespace refers to the destination
	header    string
	srcValue  interface{}
	dstValue  interface{}
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

func AsDecodingError(err error) *DecodingError {
	if err == nil {
		return nil
	}
	if e, ok := err.(*DecodingError); ok {
		return e
	}
	return NewDecodingErrorWrap(err)
}

func (e *DecodingError) SetHeader(format string, args ...interface{}) *DecodingError {
	e.header = fmt.Sprintf(format, args...)
	return e
}

func (e *DecodingError) SetSrcValue(value interface{}) *DecodingError {
	e.srcValue = value
	return e
}

func (e *DecodingError) SetDstValue(value interface{}) *DecodingError {
	e.srcValue = value
	return e
}

// Duplicate() won't duplicate any wrapped error in DecodingError for it doesn't
// know how to do it without loosing the error type (i.e. via errors.New()). Same
// applies to srcValue & dstValue.
func (e *DecodingError) Duplicate() *DecodingError {
	return &DecodingError{
		namespace: *e.namespace.Duplicate(),
		error:     e.error,
		srcValue:  e.srcValue,
		dstValue:  e.dstValue,
	}
}

func (e *DecodingError) GetSrcValue() interface{} {
	return e.srcValue
}

func (e *DecodingError) GetDstValue() interface{} {
	return e.dstValue
}

func (e *DecodingError) GetNamespace() *Namespace {
	return e.namespace.Duplicate()
}

func (e *DecodingError) SetNamespace(namespace Namespace) LocalizedError {
	e.namespace = *namespace.Duplicate()
	return e
}

func (e *DecodingError) PrependNamespace(ns Namespace) LocalizedError {
	e.namespace.PrependNamespace(ns)
	return e
}

func (e *DecodingError) AppendNamespace(ns Namespace) LocalizedError {
	e.namespace.AppendNamespace(ns)
	return e
}

func (e *DecodingError) SetNamespaceUseFieldName(useFieldName bool) LocalizedError {
	e.namespace.UseFldName(useFieldName)
	return e
}

func (e *DecodingError) Error() string {
	if e.namespace.Len() > 0 {
		return fmt.Sprintf("@'%s': %s%s", &e.namespace, e.header, e.error.Error())
	}
	return e.error.Error()
}

func (e *DecodingError) Unwrap() error {
	return e.error
}

// Error implements the error interface and can represents multiple
// errors that occur in the course of a single decode.

type DecodingErrorsFormatter func(e *DecodingErrors) string

func DefaultDecodingErrorsFormatter(e *DecodingErrors) string {
	nErrors := e.Len()
	points := make([]string, nErrors)
	for i := 0; i < nErrors; i++ {
		points[i] = fmt.Sprintf("* %s", e.Get(i).Error())
	}
	sort.Strings(points)
	return fmt.Sprintf("%d error(s) decoding:\n\n%s",
		nErrors, strings.Join(points, "\n"))
}

type DecodingErrors struct {
	formatter DecodingErrorsFormatter
	errors    []DecodingError
}

func NewDecodingErrors() *DecodingErrors {
	return &DecodingErrors{}
}

func AsDecodingErrors(err error) *DecodingErrors {
	if err == nil {
		return nil
	}
	if e, ok := err.(*DecodingErrors); ok {
		return e
	}
	return NewDecodingErrors().Append(err)
}

func (e *DecodingErrors) SetFormatter(formatter DecodingErrorsFormatter) *DecodingErrors {
	e.formatter = formatter
	return e
}

func (e *DecodingErrors) Len() int {
	return len(e.errors)
}

func (e *DecodingErrors) Get(i int) *DecodingError {
	if i >= len(e.errors) {
		return nil
	}
	return e.errors[i].Duplicate()
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
	default:
		e.errors = append(e.errors, *AsDecodingError(err))
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

func (e *DecodingErrors) SetNamespace(ns Namespace) LocalizedError {
	errors := e.errors
	for i, err := range e.errors {
		errors[i] = *err.SetNamespace(ns).(*DecodingError)
	}
	return e
}

func (e *DecodingErrors) PrependNamespace(ns Namespace) LocalizedError {
	errors := e.errors
	for i, err := range e.errors {
		errors[i] = *err.PrependNamespace(ns).(*DecodingError)
	}
	return e
}

func (e *DecodingErrors) AppendNamespace(ns Namespace) LocalizedError {
	errors := e.errors
	for i, err := range e.errors {
		errors[i] = *err.AppendNamespace(ns).(*DecodingError)
	}
	return e
}

func (e *DecodingErrors) SetNamespaceUseFieldName(useFieldName bool) LocalizedError {
	for i, err := range e.errors {
		e.errors[i] = *err.SetNamespaceUseFieldName(useFieldName).(*DecodingError)
	}
	return e
}

func (e *DecodingErrors) Error() string {
	formatter := e.formatter
	if formatter == nil {
		formatter = DefaultDecodingErrorsFormatter
	}
	return formatter(e)
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
