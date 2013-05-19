package mapstructure

import (
	"errors"
	"reflect"
	"strings"
)

// MapToStruct takes a map and uses reflection to convert it into the
// given Go native structure. val must be a pointer to a struct.
func MapToStruct(m map[string]interface{}, rawVal interface{}) error {
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

		mapVal := reflect.ValueOf(rawMapVal)
		if !mapVal.IsValid() {
			// This should never happen because we got the value out
			// of the map.
			panic("map value is not valid")
		}

		field.Set(mapVal)
	}

	return nil
}
