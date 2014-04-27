package mapstructure

import (
	"reflect"
	"strings"
)

// ComposeDecodeHookFunc creates a single DecodeHookFunc that
// automatically composes multiple DecodeHookFuncs.
//
// The composed funcs are called in order, with the result of the
// previous transformation.
func ComposeDecodeHookFunc(fs ...DecodeHookFunc) DecodeHookFunc {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		var err error
		for _, f1 := range fs {
			data, err = f1(f, t, data)
			if err != nil {
				return nil, err
			}

			// Modify the from kind to be correct with the new data
			f = getKind(reflect.ValueOf(data))
		}

		return data, nil
	}
}

// StringToSliceHookFunc returns a DecodeHookFunc that converts
// string to []string by splitting on the given sep.
func StringToSliceHookFunc(sep string) DecodeHookFunc {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || t != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		return strings.Split(raw, sep), nil
	}
}
