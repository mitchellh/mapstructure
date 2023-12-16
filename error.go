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
	useTag bool
	name   string
	tag    string
}

func NewNamespaceFld(name string) *NamespaceFld {
	return &NamespaceFld{
		name: name,
	}
}

func (nf *NamespaceFld) UseTag(useTag bool) *NamespaceFld {
	nf.useTag = useTag
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
	// tag wille be used if it's defined
	if nf.useTag && nf.tag != "" {
		return nf.tag
	}
	return nf.name
}

type NamespaceFormatter func(ns Namespace) string

func NamespaceFormatterDefault(ns Namespace) string {
	var result string

	valueToStr := func(item interface{}, sep string) {
		switch value := item.(type) {
		case NamespaceFld:
			result += sep + value.String()
		case NamespaceIdx:
			result += fmt.Sprintf("[%d]", int(value))
		case NamespaceKey:
			result += fmt.Sprintf("[%v]", value)
		}
	}
	if len(ns.items) > 0 {
		valueToStr(ns.items[0], "")
	}
	for i := 1; i < len(ns.items); i++ {
		valueToStr(ns.items[i], ".")
	}
	return result
}

type Namespace struct {
	formatter NamespaceFormatter
	items     []interface{}
}

func NewNamespace() *Namespace {
	return &Namespace{
		formatter: NamespaceFormatterDefault,
	}
}

func (ns *Namespace) SetFormatter(formatter NamespaceFormatter) *Namespace {
	if formatter != nil {
		ns.formatter = formatter
	}
	return ns
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

// AppendFldName() just calls AppendFld(): makes the syntax a little lighter
func (ns *Namespace) AppendFldName(fldNames ...string) *Namespace {
	for _, fn := range fldNames {
		ns.items = append(ns.items, *NewNamespaceFld(fn))
	}
	return ns
}

// PrependFldName() just calls PrependFld(): makes the syntax a little lighter
func (ns *Namespace) PrependFldName(fldNames ...string) *Namespace {
	ns.items = append(NewNamespace().AppendFldName(fldNames...).items, ns.items...)
	return ns
}

// UseFldTag() set preference on using the tag in place of the field name
// when the former it's defined. It affects the vale returned by String()
func (ns *Namespace) UseFldTag(useFldTag bool) *Namespace {
	for i, item := range ns.items {
		if fld, ok := item.(NamespaceFld); ok {
			ns.items[i] = *fld.UseTag(useFldTag)
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

func (ns *Namespace) String() string {
	return ns.formatter(*ns)
}

func (ns *Namespace) Duplicate() *Namespace {
	ns_ := *ns
	ns_.items = ns.items[:]
	return &ns_
}

type LocalizedError interface {
	SetNamespace(ns Namespace) LocalizedError
	PrependNamespace(ns Namespace) LocalizedError
	AppendNamespace(ns Namespace) LocalizedError
	SetNamespaceUseFldTag(useFieldTag bool) LocalizedError
	Error() string
}

func AsLocalizedError(err error) LocalizedError {
	if e, ok := err.(LocalizedError); ok {
		return e
	}
	return AsDecodingError(err)
}

type DecodingErrorKind int

const (
	// destination type is not supported
	DecodingErrorUnsupportedType DecodingErrorKind = iota
	// source type is unexpected: can't be assigned/converted to destination
	DecodingErrorUnexpectedType
	// source value of a different type cannot be parsed into the destination type (WeaklyTypedInput)
	DecodingErrorParseFailure
	// source value of a different type cannot be parsed into the destination type because of overflow (WeaklyTypedInput)
	DecodingErrorParseOverflow
	// source value of a different type cannot be decoded into the destination type through encoding/json pkg
	DecodingErrorJSONDecodeFailure
	// failed to squash a source struct field into destination (field's not a struct)
	DecodingErrorSrcSquashFailure
	// failed to squash a destination struct field into destination (field's not a struct)
	DecodingErrorDstSquashFailure
	// destination value is an array and source value is of greater size
	DecodingErrorIncompatibleSize
	// when destination is a struct and source is a map some keys of which do not correspond to any struct field (with ErrorUnused flag)
	DecodingErrorUnusedKeys
	// when destination is a struct and source is a map whose keys do not cover all the struct fields (with ErrorUnset flag)
	DecodingErrorUnsetFields
	// a generic error is a not better specified one
	DecodingErrorGeneric
	// custom user error, which also marks the border between internal mapstructure errors and new possible user defined errors
	DecodingErrorCustom
)

func (k DecodingErrorKind) String() string {
	switch k {
	case DecodingErrorUnsupportedType:
		return "unsupported type"
	case DecodingErrorUnexpectedType:
		return "unexpected type"
	case DecodingErrorParseFailure:
		return "parse failure"
	case DecodingErrorJSONDecodeFailure:
		return "JSON decode failure"
	case DecodingErrorSrcSquashFailure:
		return "source squash failure"
	case DecodingErrorDstSquashFailure:
		return "destination squash failure"
	case DecodingErrorIncompatibleSize:
		return "incompatible size"
	case DecodingErrorUnusedKeys:
		return "unused keys"
	case DecodingErrorUnsetFields:
		return "unset fields"
	case DecodingErrorGeneric:
		return "generic error"
	case DecodingErrorCustom:
		fallthrough
	default:
		return fmt.Sprintf("custom(%d)", k)
	}
}

type DecodingError struct {
	namespace Namespace // namespace refers to the destination
	kind      DecodingErrorKind
	header    string
	srcValue  interface{}
	dstValue  interface{}
	error     error
}

func NewDecodingError(kind DecodingErrorKind) *DecodingError {
	return &DecodingError{
		kind:  kind,
		error: fmt.Errorf("%s", kind),
	}
}

func AsDecodingError(err error) *DecodingError {
	if err == nil {
		return nil
	}
	if e, ok := err.(*DecodingError); ok {
		return e
	}
	return NewDecodingError(DecodingErrorGeneric).Wrap(err)
}

func (e *DecodingError) Format(format string, args ...interface{}) *DecodingError {
	return &DecodingError{
		error: fmt.Errorf(format, args...),
	}
}

func (e *DecodingError) Wrap(err error) *DecodingError {
	return &DecodingError{error: err}
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
	e_ := *e
	e_.namespace = *e.namespace.Duplicate()
	return &e_
}

func (e *DecodingError) IsCustom() bool {
	return e.kind >= DecodingErrorCustom
}

func (e *DecodingError) GetKind() DecodingErrorKind {
	return e.kind
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

func (e *DecodingError) SetNamespaceUseFldTag(useFldTag bool) LocalizedError {
	e.namespace.UseFldTag(useFldTag)
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
	nErrors := len(e.errors)
	points := make([]string, nErrors)
	for i := 0; i < nErrors; i++ {
		points[i] = fmt.Sprintf("* %s", e.errors[i].Error())
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

func (e *DecodingErrors) SetNamespaceUseFldTag(useFldTag bool) LocalizedError {
	for i, err := range e.errors {
		e.errors[i] = *err.SetNamespaceUseFldTag(useFldTag).(*DecodingError)
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
