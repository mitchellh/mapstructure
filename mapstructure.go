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
	"sort"
	"strconv"
	"strings"
)

// DecoderConfig is the configuration that is used to create a new decoder
// and allows customization of various aspects of decoding.
type DecoderConfig struct {
	// If ErrorUnused is true, then it is an error for there to exist
	// keys in the original map that were unused in the decoding process
	// (extra keys).
	ErrorUnused bool

	// If WeaklyTypedInput is true, the decoder will convert values between
	// the following types:
	//
	//    - numbers and bools
	//    - strings and numbers
	//    - strings and bools
	//    - empty arrays/slices and empty maps
	WeaklyTypedInput bool

	// Metadata is the struct that will contain extra metadata about
	// the decoding. If this is nil, then no metadata will be tracked.
	Metadata *Metadata

	// Result is a pointer to the struct that will contain the decoded
	// value.
	Result interface{}

	// The tag name that mapstructure reads for field names. This
	// defaults to "mapstructure"
	TagName string
}

// A Decoder takes a raw interface value and turns it into structured
// data, keeping track of rich error information along the way in case
// anything goes wrong. Unlike the basic top-level Decode method, you can
// more finely control how the Decoder behaves using the DecoderConfig
// structure. The top-level Decode method is just a convenience that sets
// up the most basic Decoder.
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
		Result:   rawVal,
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

	if config.Metadata != nil {
		if config.Metadata.Keys == nil {
			config.Metadata.Keys = make([]string, 0)
		}

		if config.Metadata.Unused == nil {
			config.Metadata.Unused = make([]string, 0)
		}
	}

	if config.TagName == "" {
		config.TagName = "mapstructure"
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

	var err error
	dataKind := d.getKind(val)

	switch dataKind {
	case reflect.Bool:
		err = d.decodeBool(name, data, val)
	case reflect.Interface:
		err = d.decodeBasic(name, data, val)
	case reflect.String:
		err = d.decodeString(name, data, val)
	case reflect.Int:
		err = d.decodeInt(name, data, val)
	case reflect.Uint:
		err = d.decodeUint(name, data, val)
	case reflect.Float32:
		err = d.decodeFloat(name, data, val)
	case reflect.Struct:
		err = d.decodeStruct(name, data, val)
	case reflect.Map:
		err = d.decodeMap(name, data, val)
	case reflect.Slice:
		err = d.decodeSlice(name, data, val)
	default:
		// If we reached this point then we weren't able to decode it
		return fmt.Errorf("%s: unsupported type: %s", name, dataKind)
	}

	// If we reached here, then we successfully decoded SOMETHING, so
	// mark the key as used if we're tracking metadata.
	if d.config.Metadata != nil && name != "" {
		d.config.Metadata.Keys = append(d.config.Metadata.Keys, name)
	}

	return err
}

func (d *Decoder) getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()

	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
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

func (d *Decoder) decodeString(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)

	switch {
	case dataKind == reflect.String:
		val.SetString(dataVal.String())
	case dataKind == reflect.Bool && d.config.WeaklyTypedInput:
		if dataVal.Bool() {
			val.SetString("1")
		} else {
			val.SetString("0")
		}
	case dataKind == reflect.Int && d.config.WeaklyTypedInput:
		val.SetString(strconv.FormatInt(dataVal.Int(), 10))
	case dataKind == reflect.Uint && d.config.WeaklyTypedInput:
		val.SetString(strconv.FormatUint(dataVal.Uint(), 10))
	case dataKind == reflect.Float32 && d.config.WeaklyTypedInput:
		val.SetString(strconv.FormatFloat(dataVal.Float(), 'f', -1, 64))
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s'",
			name, val.Type(), dataVal.Type())
	}

	return nil
}

func (d *Decoder) decodeInt(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)

	switch {
	case dataKind == reflect.Int:
		val.SetInt(dataVal.Int())
	case dataKind == reflect.Uint:
		val.SetInt(int64(dataVal.Uint()))
	case dataKind == reflect.Bool && d.config.WeaklyTypedInput:
		if dataVal.Bool() {
			val.SetInt(1)
		} else {
			val.SetInt(0)
		}
	case dataKind == reflect.Float32:
		val.SetInt(int64(dataVal.Float()))
	case dataKind == reflect.String && d.config.WeaklyTypedInput:
		i, err := strconv.ParseInt(dataVal.String(), 0, val.Type().Bits())
		if err == nil {
			val.SetInt(i)
		} else {
			return fmt.Errorf("cannot parse '%s' as int: %s", name, err)
		}
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s'",
			name, val.Type(), dataVal.Type())
	}

	return nil
}

func (d *Decoder) decodeUint(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)

	switch {
	case dataKind == reflect.Int:
		val.SetUint(uint64(dataVal.Int()))
	case dataKind == reflect.Uint:
		val.SetUint(dataVal.Uint())
	case dataKind == reflect.Bool && d.config.WeaklyTypedInput:
		if dataVal.Bool() {
			val.SetUint(1)
		} else {
			val.SetUint(0)
		}
	case dataKind == reflect.Float32:
		val.SetUint(uint64(dataVal.Float()))
	case dataKind == reflect.String && d.config.WeaklyTypedInput:
		i, err := strconv.ParseUint(dataVal.String(), 0, val.Type().Bits())
		if err == nil {
			val.SetUint(i)
		} else {
			return fmt.Errorf("cannot parse '%s' as uint: %s", name, err)
		}
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s'",
			name, val.Type(), dataVal.Type())
	}

	return nil
}

func (d *Decoder) decodeBool(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)

	switch {
	case dataKind == reflect.Bool:
		val.SetBool(dataVal.Bool())
	case dataKind == reflect.Int && d.config.WeaklyTypedInput:
		val.SetBool(dataVal.Int() != 0)
	case dataKind == reflect.Uint && d.config.WeaklyTypedInput:
		val.SetBool(dataVal.Uint() != 0)
	case dataKind == reflect.Float32 && d.config.WeaklyTypedInput:
		val.SetBool(dataVal.Float() != 0)
	case dataKind == reflect.String && d.config.WeaklyTypedInput:
		b, err := strconv.ParseBool(dataVal.String())
		if err == nil {
			val.SetBool(b)
		} else if dataVal.String() == "" {
			val.SetBool(false)
		} else {
			return fmt.Errorf("cannot parse '%s' as bool: %s", name, err)
		}
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s'",
			name, val.Type(), dataVal.Type())
	}

	return nil
}

func (d *Decoder) decodeFloat(name string, data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	dataKind := d.getKind(dataVal)

	switch {
	case dataKind == reflect.Int:
		val.SetFloat(float64(dataVal.Int()))
	case dataKind == reflect.Uint:
		val.SetFloat(float64(dataVal.Uint()))
	case dataKind == reflect.Bool && d.config.WeaklyTypedInput:
		if dataVal.Bool() {
			val.SetFloat(1)
		} else {
			val.SetFloat(0)
		}
	case dataKind == reflect.Float32:
		val.SetFloat(float64(dataVal.Float()))
	case dataKind == reflect.String && d.config.WeaklyTypedInput:
		f, err := strconv.ParseFloat(dataVal.String(), val.Type().Bits())
		if err == nil {
			val.SetFloat(f)
		} else {
			return fmt.Errorf("cannot parse '%s' as float: %s", name, err)
		}
	default:
		return fmt.Errorf(
			"'%s' expected type '%s', got unconvertible type '%s'",
			name, val.Type(), dataVal.Type())
	}

	return nil
}

func (d *Decoder) decodeMap(name string, data interface{}, val reflect.Value) error {
	valType := val.Type()
	valKeyType := valType.Key()
	valElemType := valType.Elem()

	// Make a new map to hold our result
	mapType := reflect.MapOf(valKeyType, valElemType)
	valMap := reflect.MakeMap(mapType)

	// Check input type
	dataVal := reflect.Indirect(reflect.ValueOf(data))
	if dataVal.Kind() != reflect.Map {
		// Accept empty array/slice instead of an empty map in weakly typed mode
		if d.config.WeaklyTypedInput &&
			(dataVal.Kind() == reflect.Slice || dataVal.Kind() == reflect.Array) &&
			dataVal.Len() == 0 {
			val.Set(valMap)
			return nil
		} else {
			return fmt.Errorf("'%s' expected a map, got '%s'", name, dataVal.Kind())
		}
	}

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
	valType := val.Type()
	valElemType := valType.Elem()

	// Make a new slice to hold our result, same size as the original data.
	sliceType := reflect.SliceOf(valElemType)
	valSlice := reflect.MakeSlice(sliceType, dataVal.Len(), dataVal.Len())

	// Check input type
	if dataValKind != reflect.Array && dataValKind != reflect.Slice {
		// Accept empty map instead of array/slice in weakly typed mode
		if d.config.WeaklyTypedInput && dataVal.Kind() == reflect.Map && dataVal.Len() == 0 {
			val.Set(valSlice)
			return nil
		} else {
			return fmt.Errorf(
				"'%s': source data must be an array or slice, got %s", name, dataValKind)
		}
	}

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

	dataValKeys := make(map[reflect.Value]struct{})
	dataValKeysUnused := make(map[interface{}]struct{})
	for _, dataValKey := range dataVal.MapKeys() {
		dataValKeys[dataValKey] = struct{}{}
		dataValKeysUnused[dataValKey.Interface()] = struct{}{}
	}

	// This slice will keep track of all the structs we'll be decoding.
	// There can be more than one struct if there are embedded structs
	// that are squashed.
	structs := make([]reflect.Value, 1, 5)
	structs[0] = val

	// Compile the list of all the fields that we're going to be decoding
	// from all the structs.
	fields := make(map[*reflect.StructField]reflect.Value)
	for len(structs) > 0 {
		structVal := structs[0]
		structs = structs[1:]

		structType := structVal.Type()
		for i := 0; i < structType.NumField(); i++ {
			fieldType := structType.Field(i)

			if fieldType.Anonymous {
				// We have an embedded field. We "squash" the fields down
				// if specified in the tag.
				squash := false
				tagParts := strings.Split(fieldType.Tag.Get(d.config.TagName), ",")
				for _, tag := range tagParts[1:] {
					if tag == "squash" {
						squash = true
						break
					}
				}

				if squash {
					structs = append(structs, val.FieldByName(fieldType.Name))
					continue
				}
			}

			// Normal struct field, store it away
			fields[&fieldType] = structVal.Field(i)
		}
	}

	errors := make([]string, 0)
	for fieldType, field := range fields {
		fieldName := fieldType.Name

		tagValue := fieldType.Tag.Get(d.config.TagName)
		tagValue = strings.SplitN(tagValue, ",", 2)[0]
		if tagValue != "" {
			fieldName = tagValue
		}

		rawMapKey := reflect.ValueOf(fieldName)
		rawMapVal := dataVal.MapIndex(rawMapKey)
		if !rawMapVal.IsValid() {
			// Do a slower search by iterating over each key and
			// doing case-insensitive search.
			for dataValKey, _ := range dataValKeys {
				mK := dataValKey.Interface().(string)

				if strings.EqualFold(mK, fieldName) {
					rawMapKey = dataValKey
					rawMapVal = dataVal.MapIndex(dataValKey)
					break
				}
			}

			if !rawMapVal.IsValid() {
				// There was no matching key in the map for the value in
				// the struct. Just ignore.
				continue
			}
		}

		// Delete the key we're using from the unused map so we stop tracking
		delete(dataValKeysUnused, rawMapKey.Interface())

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

	if d.config.ErrorUnused && len(dataValKeysUnused) > 0 {
		keys := make([]string, 0, len(dataValKeysUnused))
		for rawKey, _ := range dataValKeysUnused {
			keys = append(keys, rawKey.(string))
		}
		sort.Strings(keys)

		err := fmt.Errorf("'%s' has invalid keys: %s", name, strings.Join(keys, ", "))
		errors = appendErrors(errors, err)
	}

	if len(errors) > 0 {
		return &Error{errors}
	}

	// Add the unused keys to the list of unused keys if we're tracking metadata
	if d.config.Metadata != nil {
		for rawKey, _ := range dataValKeysUnused {
			key := rawKey.(string)
			if name != "" {
				key = fmt.Sprintf("%s.%s", name, key)
			}

			d.config.Metadata.Unused = append(d.config.Metadata.Unused, key)
		}
	}

	return nil
}
