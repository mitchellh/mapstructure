package mapstructure

import (
	"strconv"
)

//PathPart is interface for different kinds of FieldPath elements.
type PathPart interface {
	getDelimiter() string
	String() string
}

//InStructPathPart is FieldPath element that represents field name in structure.
type InStructPathPart struct {
	val string
}

func (p InStructPathPart) getDelimiter() string {
	return "."
}

func (p InStructPathPart) String() string {
	return p.val
}

func (p InStructPathPart) Value() string {
	return p.val
}

//InMapPathPart is FieldPath element that represents key in map.
type InMapPathPart struct {
	val string
}

func (p InMapPathPart) getDelimiter() string {
	return ""
}

func (p InMapPathPart) String() string {
	return "[" + p.val + "]"
}

func (p InMapPathPart) Value() string {
	return p.val
}

//InSlicePathPart is FieldPath element that represents index in slice or array.
type InSlicePathPart struct {
	val int
}

func (p InSlicePathPart) getDelimiter() string {
	return ""
}

func (p InSlicePathPart) String() string {
	return "[" + strconv.Itoa(p.val) + "]"
}

func (p InSlicePathPart) Value() int {
	return p.val
}

//FieldPath represents path to a field in nested structure.
type FieldPath struct {
	parts []PathPart
}

func (f FieldPath) addStruct(part string) FieldPath {
	return FieldPath{
		parts: appendPart(f.parts, InStructPathPart{val: part}),
	}
}

func (f FieldPath) addMap(part string) FieldPath {
	return FieldPath{
		parts: appendPart(f.parts, InMapPathPart{val: part}),
	}
}

func (f FieldPath) addSlice(part int) FieldPath {
	return FieldPath{
		parts: appendPart(f.parts, InSlicePathPart{val: part}),
	}
}

func (f FieldPath) notEmpty() bool {
	return len(f.parts) > 0
}

func newFieldPath() FieldPath {
	return FieldPath{
		parts: make([]PathPart, 0),
	}
}

func (f FieldPath) Parts() []PathPart {
	return f.parts
}

func (f FieldPath) String() string {
	result := ""

	for i, part := range f.parts {
		delimiter := ""

		if i > 0 { //there is no delimiter before first element
			delimiter = part.getDelimiter()
		}

		result += delimiter + part.String()
	}

	return result
}

//appendPart appends PathPart to a PathPart slice with guarantee of slice immutability.
func appendPart(parts []PathPart, part PathPart) []PathPart {
	p := make([]PathPart, len(parts))
	copy(p, parts)
	return append(p, part)
}
