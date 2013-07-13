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

// DecoderConfig is the configuration that is used to create a new decoder
// and allows customization of various aspects of decoding.
type DecoderConfig struct {
	// Metadata is the struct that will contain extra metadata about
	// the decoding. If this is nil, then no metadata will be tracked.
	Metadata *Metadata

	// Result is a pointer to the struct that will contain the decoded
	// value.
	Result interface{}
}

// A Decoder takes a raw interface value and turns it into structured
// data, keeping track of rich error information along the way in case
// anything goes wrong.
type Decoder struct {
	config *DecoderConfig
}

// Metadata contains information about decoding a structure that
// is tedious or difficult to get otherwise.
type Metadata struct {
	// Keys are the keys of the structure which were successfully decoded
	Keys []string

	// Unused is a slice of keys that were found in the raw value but
	// weren't decoded since there was no matching field in the result interface
	Unused []string
}

// Decode takes a map and uses reflection to convert it into the
// given Go native structure. val must be a pointer to a struct.
func Decode(m interface{}, rawVal interface{}) error {
	config := &DecoderConfig{
		Metadata: nil,
		Result: rawVal,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(m)
}

// NewDecoder returns a new decoder for the given configuration. Once
// a decoder has been returned, the same configuration must not be used
// again.
func NewDecoder(config *DecoderConfig) (*Decoder, error) {
	val := reflect.ValueOf(config.Result)
	if val.Kind() != reflect.Ptr {
		return nil, errors.New("result must be a pointer")
	}

	val = val.Elem()
	if !val.CanAddr() {
		return nil, errors.New("result must be addressable (a pointer)")
	}

	result := &Decoder{
		config: config,
	}

	return result, nil
}

// Decode decodes the given raw interface to the target pointer specified
// by the configuration.
func (d *Decoder) Decode(raw interface{}) error {
	return d.decode("", raw, reflect.ValueOf(d.config.Result).Elem())
}

// Decodes an unknown data type into a specific reflection value.
func (d *Decoder) decode(name string, data interface{}, val reflect.Value) error {
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
		return d.decodeBasic(name, data, val)
	case reflect.Int:
		fallthrough
	case reflect.Uint:
		return d.decodeInt(name, data, val)
	case reflect.Struct:
		return d.decodeStruct(name, data, val)
	case reflect.Map:
		return d.decodeMap(name, data, val)
	case reflect.Slice:
		return d.decodeSlice(name, data, val)
	}

	// If we reached this point then we weren't able to decode it
	return fmt.Errorf("%s: unsupported type: %s", name, k)
}

// This decodes a basic type (bool, int, string, etc.) and sets the
// value to "data" of that type.
func (d *Decoder) decodeBasic(name string, data interface{}, val reflect.Value) error {
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

func (d *Decoder) decodeInt(name string, data interface{}, val reflect.Value) error {
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

func (d *Decoder) decodeMap(name string, data interface{}, val reflect.Value) error {
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
	errors := make([]string, 0)

	for _, k := range dataVal.MapKeys() {
		fieldName := fmt.Sprintf("%s[%s]", name, k)

		// First decode the key into the proper type
		currentKey := reflect.Indirect(reflect.New(valKeyType))
		if err := d.decode(fieldName, k.Interface(), currentKey); err != nil {
			errors = appendErrors(errors, err)
			continue
		}

		// Next decode the data into the proper type
		v := dataVal.MapIndex(k).Interface()
		currentVal := reflect.Indirect(reflect.New(valElemType))
		if err := d.decode(fieldName, v, currentVal); err != nil {
			errors = appendErrors(errors, err)
			continue
		}

		valMap.SetMapIndex(currentKey, currentVal)
	}

	// Set the built up map to the value
	val.Set(valMap)

	// If we had errors, return those
	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}

func (d *Decoder) decodeSlice(name string, data interface{}, val reflect.Value) error {
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
	errors := make([]string, 0)

	for i := 0; i < dataVal.Len(); i++ {
		currentData := dataVal.Index(i).Interface()
		currentField := valSlice.Index(i)

		fieldName := fmt.Sprintf("%s[%d]", name, i)
		if err := d.decode(fieldName, currentData, currentField); err != nil {
			errors = appendErrors(errors, err)
		}
	}

	// Finally, set the value to the slice we built up
	val.Set(valSlice)

	// If there were errors, we return those
	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}

func (d *Decoder) decodeStruct(name string, data interface{}, val reflect.Value) error {
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

	errors := make([]string, 0)

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

		if err := d.decode(fieldName, rawMapVal.Interface(), field); err != nil {
			errors = appendErrors(errors, err)
		}
	}

	if len(errors) > 0 {
		return &Error{errors}
	}

	return nil
}
