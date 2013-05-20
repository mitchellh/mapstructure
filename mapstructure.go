package mapstructure

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Decode takes a map and uses reflection to convert it into the
// given Go native structure. val must be a pointer to a struct.
func Decode(m map[string]interface{}, rawVal interface{}) error {
	val := reflect.ValueOf(rawVal)
	if val.Kind() != reflect.Ptr {
		return errors.New("val must be a pointer")
	}

	val = val.Elem()
	if !val.CanAddr() {
		return errors.New("val must be addressable (a pointer)")
	}

	if val.Kind() != reflect.Struct {
		return errors.New("val must be an addressable struct")
	}

	return decode("root", m, val)
}

// Decodes an unknown data type into a specific reflection value.
func decode(name string, data interface{}, val reflect.Value) error {
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
	case reflect.Int:
		fallthrough
	case reflect.String:
		fallthrough
	case reflect.Uint:
		return decodeBasic(name, data, val)
	case reflect.Struct:
		return decodeStruct(name, data, val)
	case reflect.Map:
		return decodeMap(name, data, val)
	case reflect.Slice:
		return decodeSlice(name, data, val)
	}

	// If we reached this point then we weren't able to decode it
	return fmt.Errorf("unsupported type: %s", k)
}

// This decodes a basic type (bool, int, string, etc.) and sets the
// value to "data" of that type.
func decodeBasic(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	if !dataVal.IsValid() {
		// This should never happen because upstream makes sure it is valid
		panic("data is invalid")
	}

	dataValType := dataVal.Type()
	if !dataValType.AssignableTo(val.Type()) {
		return fmt.Errorf(
			"'%s' expected type '%s', got '%s'",
			name, val.Type(), dataValType)
	}

	val.Set(dataVal)
	return nil
}

func decodeMap(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	if dataVal.Kind() != reflect.Map {
		return fmt.Errorf("'%s' expected a map, got '%s'", name, dataVal.Kind())
	}

	dataValType := dataVal.Type()
	if dataValType.Key().Kind() != reflect.String {
		return fmt.Errorf(
			"'%s' needs a map with string keys, has '%s' keys",
			name, dataValType.Key().Kind())
	}

	// Just go ahead and set one map to the other...
	val.Set(dataVal)

	return nil
}

func decodeSlice(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	val.Set(dataVal)
	return nil
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

	// At this point we know that data is a map with string keys, so
	// we can properly cast it here. We use the "Interface()" value because
	// this gets us the proper interface whether or not data is a pointer
	// or not.
	m, ok := dataVal.Interface().(map[string]interface{})
	if !ok {
		panic("data could not be cast as map[string]interface{}")
	}

	valType := val.Type()
	for i := 0; i < valType.NumField(); i++ {
		fieldType := valType.Field(i)
		fieldName := fieldType.Name

		rawMapVal, ok := m[fieldName]
		if !ok {
			// Do a slower search by iterating over each key and
			// doing case-insensitive search.
			for mK, mV := range m {
				if strings.EqualFold(mK, fieldName) {
					rawMapVal = mV
					break
				}
			}

			if rawMapVal == nil {
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

		fieldName = fmt.Sprintf("%s.%s", name, fieldName)
		if err := decode(fieldName, rawMapVal, field); err != nil {
			return err
		}
	}

	return nil
}
