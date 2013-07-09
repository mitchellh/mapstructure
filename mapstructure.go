// The mapstructure package exposes functionality to convert an
// abitrary map[string]interface{} into a native Go structure.
//
// The Go structure can be arbitrarily complex, containing slices,
// other structs, etc. and the decoder will properly decode nested
// maps and so on into the proper structures in the native Go struct.
// See the examples to see what the decoder is capable of.
package mapstructure

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ErrorAccumulator accumulates error messages during a single decode.
type ErrorAccumulator []string

func (e ErrorAccumulator) append(err error) ErrorAccumulator {
	if es, ok := err.(Error); ok {
		return append(e, es...)
	}
	return append(e, err.Error())
}

func (e ErrorAccumulator) flush() error {
	if len(e) == 0 {
		return nil
	}
	return Error(e)
}

// Error represents multiple errors that occur in the course
// of a single decode.
type Error []string

func (e Error) Error() string {
	return fmt.Sprintf(
		"%d error(s) decoding:\n\n%s",
		len(e), "* "+strings.Join([]string(e), "\n* "))
}

// Decode takes a map and uses reflection to convert it into the
// given Go native structure. val must be a pointer to a struct.
func Decode(m interface{}, rawVal interface{}) error {
	val := reflect.ValueOf(rawVal)
	if val.Kind() != reflect.Ptr {
		return errors.New("val must be a pointer")
	}

	val = val.Elem()
	if !val.CanAddr() {
		return errors.New("val must be addressable (a pointer)")
	}

	return decode("", m, val)
}

// Decodes an unknown data type into a specific reflection value.
func decode(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	if !dataVal.IsValid() {
		// If the data value is invalid, then we just set the value
		// to be the zero value.
		val.Set(reflect.Zero(val.Type()))
		return nil
	}

	k := val.Kind()

	// Some shortcuts because we treat all ints and uints the same way
	if k >= reflect.Int && k <= reflect.Int64 {
		k = reflect.Int
	} else if k >= reflect.Uint && k <= reflect.Uint64 {
		k = reflect.Uint
	}

	switch k {
	case reflect.Bool:
		fallthrough
	case reflect.Interface:
		fallthrough
	case reflect.String:
		return decodeBasic(name, data, val)
	case reflect.Int:
		fallthrough
	case reflect.Uint:
		return decodeInt(name, data, val)
	case reflect.Struct:
		return decodeStruct(name, data, val)
	case reflect.Map:
		return decodeMap(name, data, val)
	case reflect.Slice:
		return decodeSlice(name, data, val)
	}

	// If we reached this point then we weren't able to decode it
	return fmt.Errorf("%s: unsupported type: %s", name, k)
}

// This decodes a basic type (bool, int, string, etc.) and sets the
// value to "data" of that type.
func decodeBasic(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataValType := dataVal.Type()
	if !dataValType.AssignableTo(val.Type()) {
		return fmt.Errorf(
			"'%s' expected type '%s', got '%s'",
			name, val.Type(), dataValType)
	}

	val.Set(dataVal)
	return nil
}

func decodeInt(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := dataVal.Kind()
	if dataKind >= reflect.Int && dataKind <= reflect.Int64 {
		dataKind = reflect.Int
	} else if dataKind >= reflect.Uint && dataKind <= reflect.Uint64 {
		dataKind = reflect.Uint
	} else if dataKind >= reflect.Float32 && dataKind <= reflect.Float64 {
		dataKind = reflect.Float32
	} else {
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s'",
			name, val.Type(), dataVal.Type())
	}

	valKind := val.Kind()
	if valKind >= reflect.Int && valKind <= reflect.Int64 {
		valKind = reflect.Int
	} else if valKind >= reflect.Uint && valKind <= reflect.Uint64 {
		valKind = reflect.Uint
	}

	switch dataKind {
	case reflect.Int:
		if valKind == reflect.Int {
			val.SetInt(dataVal.Int())
		} else {
			val.SetUint(uint64(dataVal.Int()))
		}
	case reflect.Uint:
		if valKind == reflect.Int {
			val.SetInt(int64(dataVal.Uint()))
		} else {
			val.SetUint(dataVal.Uint())
		}
	case reflect.Float32:
		if valKind == reflect.Int {
			val.SetInt(int64(dataVal.Float()))
		} else {
			val.SetUint(uint64(dataVal.Float()))
		}
	default:
		panic("should never reach")
	}

	return nil
}

func decodeMap(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	if dataVal.Kind() != reflect.Map {
		return fmt.Errorf("'%s' expected a map, got '%s'", name, dataVal.Kind())
	}

	valType := val.Type()
	valKeyType := valType.Key()
	valElemType := valType.Elem()

	// Make a new map to hold our result
	mapType := reflect.MapOf(valKeyType, valElemType)
	valMap := reflect.MakeMap(mapType)

	// Accumulate errors
	var errors ErrorAccumulator = nil

	for _, k := range dataVal.MapKeys() {
		fieldName := fmt.Sprintf("%s[%s]", name, k)

		// First decode the key into the proper type
		currentKey := reflect.Indirect(reflect.New(valKeyType))
		if err := decode(fieldName, k.Interface(), currentKey); err != nil {
			errors = errors.append(err)
			continue
		}

		// Next decode the data into the proper type
		v := dataVal.MapIndex(k).Interface()
		currentVal := reflect.Indirect(reflect.New(valElemType))
		if err := decode(fieldName, v, currentVal); err != nil {
			errors = errors.append(err)
			continue
		}

		valMap.SetMapIndex(currentKey, currentVal)
	}

	// Set the built up map to the value
	val.Set(valMap)

	return errors.flush()
}

func decodeSlice(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataValKind := dataVal.Kind()
	if dataValKind != reflect.Array && dataValKind != reflect.Slice {
		return fmt.Errorf(
			"'%s': source data must be an array or slice, got %s", name, dataValKind)
	}

	valType := val.Type()
	valElemType := valType.Elem()

	// Make a new slice to hold our result, same size as the original data.
	sliceType := reflect.SliceOf(valElemType)
	valSlice := reflect.MakeSlice(sliceType, dataVal.Len(), dataVal.Len())

	// Accumulate any errors
	var errors ErrorAccumulator = nil

	for i := 0; i < dataVal.Len(); i++ {
		currentData := dataVal.Index(i).Interface()
		currentField := valSlice.Index(i)

		fieldName := fmt.Sprintf("%s[%d]", name, i)
		if err := decode(fieldName, currentData, currentField); err != nil {
			errors = errors.append(err)
		}
	}

	// Finally, set the value to the slice we built up
	val.Set(valSlice)

	return errors.flush()
}

func decodeStruct(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	dataValKind := dataVal.Kind()
	if dataValKind != reflect.Map {
		return fmt.Errorf("'%s' expected a map, got '%s'", name, dataValKind)
	}

	dataValType := dataVal.Type()
	if dataValType.Key().Kind() != reflect.String {
		return fmt.Errorf(
			"'%s' needs a map with string keys, has '%s' keys",
			name, dataValType.Key().Kind())
	}

	var errors ErrorAccumulator = nil

	valType := val.Type()
	for i := 0; i < valType.NumField(); i++ {
		fieldType := valType.Field(i)
		fieldName := fieldType.Name

		tagValue := fieldType.Tag.Get("mapstructure")
		if tagValue != "" {
			fieldName = tagValue
		}

		rawMapVal := dataVal.MapIndex(reflect.ValueOf(fieldName))
		if !rawMapVal.IsValid() {
			// Do a slower search by iterating over each key and
			// doing case-insensitive search.
			for _, dataKeyVal := range dataVal.MapKeys() {
				mK := dataKeyVal.Interface().(string)

				if strings.EqualFold(mK, fieldName) {
					rawMapVal = dataVal.MapIndex(dataKeyVal)
					break
				}
			}

			if !rawMapVal.IsValid() {
				// There was no matching key in the map for the value in
				// the struct. Just ignore.
				continue
			}
		}

		field := val.Field(i)
		if !field.IsValid() {
			// This should never happen
			panic("field is not valid")
		}

		// If we can't set the field, then it is unexported or something,
		// and we just continue onwards.
		if !field.CanSet() {
			continue
		}

		// If the name is empty string, then we're at the root, and we
		// don't dot-join the fields.
		if name != "" {
			fieldName = fmt.Sprintf("%s.%s", name, fieldName)
		}

		if err := decode(fieldName, rawMapVal.Interface(), field); err != nil {
			errors = errors.append(err)
		}
	}

	return errors.flush()
}
