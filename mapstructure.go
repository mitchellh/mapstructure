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

	valType := val.Type()
	valKeyType := valType.Key()
	valElemType := valType.Elem()

	// Make a new map to hold our result
	mapType := reflect.MapOf(valKeyType, valElemType)
	valMap := reflect.MakeMap(mapType)

	for _, k := range dataVal.MapKeys() {
		currentData := dataVal.MapIndex(k).Interface()
		currentVal := reflect.Indirect(reflect.New(valElemType))

		fieldName := fmt.Sprintf("%s[%s]", name, k)
		if err := decode(fieldName, currentData, currentVal); err != nil {
			return err
		}

		valMap.SetMapIndex(k, currentVal)
	}

	val.Set(valMap)

	return nil
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

	// TODO: Error checking to make sure data is an array/slice type

	// Make a new slice to hold our result, same size as the original data.
	sliceType := reflect.SliceOf(valElemType)
	valSlice := reflect.MakeSlice(sliceType, dataVal.Len(), dataVal.Len())

	for i := 0; i < dataVal.Len(); i++ {
		currentData := dataVal.Index(i).Interface()
		currentField := valSlice.Index(i)

		fieldName := fmt.Sprintf("%s[%d]", name, i)
		if err := decode(fieldName, currentData, currentField); err != nil {
			return err
		}
	}

	// Finally, set the value to the slice we built up
	val.Set(valSlice)

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

	valType := val.Type()
	for i := 0; i < valType.NumField(); i++ {
		fieldType := valType.Field(i)
		fieldName := fieldType.Name

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

		fieldName = fmt.Sprintf("%s.%s", name, fieldName)
		if err := decode(fieldName, rawMapVal.Interface(), field); err != nil {
			return err
		}
	}

	return nil
}
