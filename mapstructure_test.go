package mapstructure

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

type Basic struct {
	Vstring string
	Vint    int
	Vuint   uint
	Vbool   bool
	Vfloat  float64
	Vextra  string
	vsilent bool
	Vdata   interface{}
}

type Embedded struct {
	Basic
	Vunique string
}

type EmbeddedPointer struct {
	*Basic
	Vunique string
}

type EmbeddedSquash struct {
	Basic   `mapstructure:",squash"`
	Vunique string
}

type Map struct {
	Vfoo   string
	Vother map[string]string
}

type MapOfStruct struct {
	Value map[string]Basic
}

type Nested struct {
	Vfoo string
	Vbar Basic
}

type Slice struct {
	Vfoo string
	Vbar []string
}

type SliceOfStruct struct {
	Value []Basic
}

type Tagged struct {
	Extra string `mapstructure:"bar,what,what"`
	Value string `mapstructure:"foo"`
}

type TypeConversionResult struct {
	IntToFloat    float32
	IntToUint     uint
	IntToBool     bool
	IntToString   string
	UintToInt     int
	UintToFloat   float32
	UintToBool    bool
	UintToString  string
	BoolToInt     int
	BoolToUint    uint
	BoolToFloat   float32
	BoolToString  string
	FloatToInt    int
	FloatToUint   uint
	FloatToBool   bool
	FloatToString string
	StringToInt   int
	StringToUint  uint
	StringToBool  bool
	StringToFloat float32
	SliceToMap    map[string]interface{}
	MapToSlice    []interface{}
}

func TestBasicTypes(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vstring": "foo",
		"vint":    42,
		"Vuint":   42,
		"vbool":   true,
		"Vfloat":  42.42,
		"vsilent": true,
		"vdata":   42,
	}

	var result Basic
	err := Decode(input, &result)
	if err != nil {
		t.Errorf("got an err: %s", err.Error())
		t.FailNow()
	}

	if result.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vstring)
	}

	if result.Vint != 42 {
		t.Errorf("vint value should be 42: %#v", result.Vint)
	}

	if result.Vuint != 42 {
		t.Errorf("vuint value should be 42: %#v", result.Vuint)
	}

	if result.Vbool != true {
		t.Errorf("vbool value should be true: %#v", result.Vbool)
	}

	if result.Vfloat != 42.42 {
		t.Errorf("vfloat value should be 42.42: %#v", result.Vfloat)
	}

	if result.Vextra != "" {
		t.Errorf("vextra value should be empty: %#v", result.Vextra)
	}

	if result.vsilent != false {
		t.Error("vsilent should not be set, it is unexported")
	}

	if result.Vdata != 42 {
		t.Error("vdata should be valid")
	}
}

func TestBasic_IntWithFloat(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vint": float64(42),
	}

	var result Basic
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}
}

func TestDecode_Embedded(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vstring": "foo",
		"Basic": map[string]interface{}{
			"vstring": "innerfoo",
		},
		"vunique": "bar",
	}

	var result Embedded
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vstring != "innerfoo" {
		t.Errorf("vstring value should be 'innerfoo': %#v", result.Vstring)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecode_EmbeddedPointer(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vstring": "foo",
		"Basic": map[string]interface{}{
			"vstring": "innerfoo",
		},
		"vunique": "bar",
	}

	var result EmbeddedPointer
	err := Decode(input, &result)
	if err == nil {
		t.Fatal("should get error")
	}
}

func TestDecode_EmbeddedSquash(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vstring": "foo",
		"vunique": "bar",
	}

	var result EmbeddedSquash
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vstring)
	}

	if result.Vunique != "bar" {
		t.Errorf("vunique value should be 'bar': %#v", result.Vunique)
	}
}

func TestDecode_DecodeHook(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vint": "WHAT",
	}

	decodeHook := func(from reflect.Kind, to reflect.Kind, v interface{}) (interface{}, error) {
		if from == reflect.String && to != reflect.String {
			return 5, nil
		}

		return v, nil
	}

	var result Basic
	config := &DecoderConfig{
		DecodeHook: decodeHook,
		Result:     &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if result.Vint != 5 {
		t.Errorf("vint should be 5: %#v", result.Vint)
	}
}

func TestDecode_Nil(t *testing.T) {
	t.Parallel()

	var input interface{} = nil
	result := Basic{
		Vstring: "foo",
	}

	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.Vstring != "foo" {
		t.Fatalf("bad: %#v", result.Vstring)
	}
}

func TestDecode_NonStruct(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"foo": "bar",
		"bar": "baz",
	}

	var result map[string]string
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result["foo"] != "bar" {
		t.Fatal("foo is not bar")
	}
}

func TestDecode_TypeConversion(t *testing.T) {
	input := map[string]interface{}{
		"IntToFloat":    42,
		"IntToUint":     42,
		"IntToBool":     1,
		"IntToString":   42,
		"UintToInt":     42,
		"UintToFloat":   42,
		"UintToBool":    42,
		"UintToString":  42,
		"BoolToInt":     true,
		"BoolToUint":    true,
		"BoolToFloat":   true,
		"BoolToString":  true,
		"FloatToInt":    42.42,
		"FloatToUint":   42.42,
		"FloatToBool":   42.42,
		"FloatToString": 42.42,
		"StringToInt":   "42",
		"StringToUint":  "42",
		"StringToBool":  "1",
		"StringToFloat": "42.42",
		"SliceToMap":    []interface{}{},
		"MapToSlice":    map[string]interface{}{},
	}

	expectedResultStrict := TypeConversionResult{
		IntToFloat:  42.0,
		IntToUint:   42,
		UintToInt:   42,
		UintToFloat: 42,
		BoolToInt:   0,
		BoolToUint:  0,
		BoolToFloat: 0,
		FloatToInt:  42,
		FloatToUint: 42,
	}

	expectedResultWeak := TypeConversionResult{
		IntToFloat:    42.0,
		IntToUint:     42,
		IntToBool:     true,
		IntToString:   "42",
		UintToInt:     42,
		UintToFloat:   42,
		UintToBool:    true,
		UintToString:  "42",
		BoolToInt:     1,
		BoolToUint:    1,
		BoolToFloat:   1,
		BoolToString:  "1",
		FloatToInt:    42,
		FloatToUint:   42,
		FloatToBool:   true,
		FloatToString: "42.42",
		StringToInt:   42,
		StringToUint:  42,
		StringToBool:  true,
		StringToFloat: 42.42,
		SliceToMap:    map[string]interface{}{},
		MapToSlice:    []interface{}{},
	}

	// Test strict type conversion
	var resultStrict TypeConversionResult
	err := Decode(input, &resultStrict)
	if err == nil {
		t.Errorf("should return an error")
	}
	if !reflect.DeepEqual(resultStrict, expectedResultStrict) {
		t.Errorf("expected %v, got: %v", expectedResultStrict, resultStrict)
	}

	// Test weak type conversion
	var decoder *Decoder
	var resultWeak TypeConversionResult

	config := &DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &resultWeak,
	}

	decoder, err = NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if !reflect.DeepEqual(resultWeak, expectedResultWeak) {
		t.Errorf("expected \n%#v, got: \n%#v", expectedResultWeak, resultWeak)
	}
}

func TestDecoder_ErrorUnused(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vstring": "hello",
		"foo":     "bar",
	}

	var result Basic
	config := &DecoderConfig{
		ErrorUnused: true,
		Result:      &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMap(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vfoo": "foo",
		"vother": map[interface{}]interface{}{
			"foo": "foo",
			"bar": "bar",
		},
	}

	var result Map
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an error: %s", err)
	}

	if result.Vfoo != "foo" {
		t.Errorf("vfoo value should be 'foo': %#v", result.Vfoo)
	}

	if result.Vother == nil {
		t.Fatal("vother should not be nil")
	}

	if len(result.Vother) != 2 {
		t.Error("vother should have two items")
	}

	if result.Vother["foo"] != "foo" {
		t.Errorf("'foo' key should be foo, got: %#v", result.Vother["foo"])
	}

	if result.Vother["bar"] != "bar" {
		t.Errorf("'bar' key should be bar, got: %#v", result.Vother["bar"])
	}
}

func TestMapOfStruct(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"value": map[string]interface{}{
			"foo": map[string]string{"vstring": "one"},
			"bar": map[string]string{"vstring": "two"},
		},
	}

	var result MapOfStruct
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err)
	}

	if result.Value == nil {
		t.Fatal("value should not be nil")
	}

	if len(result.Value) != 2 {
		t.Error("value should have two items")
	}

	if result.Value["foo"].Vstring != "one" {
		t.Errorf("foo value should be 'one', got: %s", result.Value["foo"].Vstring)
	}

	if result.Value["bar"].Vstring != "two" {
		t.Errorf("bar value should be 'two', got: %s", result.Value["bar"].Vstring)
	}
}

func TestNestedType(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vfoo": "foo",
		"vbar": map[string]interface{}{
			"vstring": "foo",
			"vint":    42,
			"vbool":   true,
		},
	}

	var result Nested
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vfoo != "foo" {
		t.Errorf("vfoo value should be 'foo': %#v", result.Vfoo)
	}

	if result.Vbar.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vbar.Vstring)
	}

	if result.Vbar.Vint != 42 {
		t.Errorf("vint value should be 42: %#v", result.Vbar.Vint)
	}

	if result.Vbar.Vbool != true {
		t.Errorf("vbool value should be true: %#v", result.Vbar.Vbool)
	}

	if result.Vbar.Vextra != "" {
		t.Errorf("vextra value should be empty: %#v", result.Vbar.Vextra)
	}
}

func TestNestedTypePointer(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vfoo": "foo",
		"vbar": &map[string]interface{}{
			"vstring": "foo",
			"vint":    42,
			"vbool":   true,
		},
	}

	var result Nested
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got an err: %s", err.Error())
	}

	if result.Vfoo != "foo" {
		t.Errorf("vfoo value should be 'foo': %#v", result.Vfoo)
	}

	if result.Vbar.Vstring != "foo" {
		t.Errorf("vstring value should be 'foo': %#v", result.Vbar.Vstring)
	}

	if result.Vbar.Vint != 42 {
		t.Errorf("vint value should be 42: %#v", result.Vbar.Vint)
	}

	if result.Vbar.Vbool != true {
		t.Errorf("vbool value should be true: %#v", result.Vbar.Vbool)
	}

	if result.Vbar.Vextra != "" {
		t.Errorf("vextra value should be empty: %#v", result.Vbar.Vextra)
	}
}

func TestSlice(t *testing.T) {
	t.Parallel()

	inputStringSlice := map[string]interface{}{
		"vfoo": "foo",
		"vbar": []string{"foo", "bar", "baz"},
	}

	inputStringSlicePointer := map[string]interface{}{
		"vfoo": "foo",
		"vbar": &[]string{"foo", "bar", "baz"},
	}

	outputStringSlice := &Slice{
		"foo",
		[]string{"foo", "bar", "baz"},
	}

	testSliceInput(t, inputStringSlice, outputStringSlice)
	testSliceInput(t, inputStringSlicePointer, outputStringSlice)
}

func TestSliceOfStruct(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"value": []map[string]interface{}{
			{"vstring": "one"},
			{"vstring": "two"},
		},
	}

	var result SliceOfStruct
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got unexpected error: %s", err)
	}

	if len(result.Value) != 2 {
		t.Fatalf("expected two values, got %d", len(result.Value))
	}

	if result.Value[0].Vstring != "one" {
		t.Errorf("first value should be 'one', got: %s", result.Value[0].Vstring)
	}

	if result.Value[1].Vstring != "two" {
		t.Errorf("second value should be 'two', got: %s", result.Value[1].Vstring)
	}
}

func TestInvalidType(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vstring": 42,
	}

	var result Basic
	err := Decode(input, &result)
	if err == nil {
		t.Fatal("error should exist")
	}

	derr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error should be kind of Error, instead: %#v", err)
	}

	if derr.Errors[0] != "'Vstring' expected type 'string', got unconvertible type 'int'" {
		t.Errorf("got unexpected error: %s", err)
	}
}

func TestMetadata(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vfoo": "foo",
		"vbar": map[string]interface{}{
			"vstring": "foo",
			"Vuint":   42,
			"foo":     "bar",
		},
		"bar": "nil",
	}

	var md Metadata
	var result Nested
	config := &DecoderConfig{
		Metadata: &md,
		Result:   &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}

	expectedKeys := []string{"Vfoo", "Vbar.Vstring", "Vbar.Vuint", "Vbar"}
	if !reflect.DeepEqual(md.Keys, expectedKeys) {
		t.Fatalf("bad keys: %#v", md.Keys)
	}

	expectedUnused := []string{"Vbar.foo", "bar"}
	if !reflect.DeepEqual(md.Unused, expectedUnused) {
		t.Fatalf("bad unused: %#v", md.Unused)
	}
}

func TestMetadata_Embedded(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vstring": "foo",
		"vunique": "bar",
	}

	var md Metadata
	var result EmbeddedSquash
	config := &DecoderConfig{
		Metadata: &md,
		Result:   &result,
	}

	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = decoder.Decode(input)
	if err != nil {
		t.Fatalf("err: %s", err.Error())
	}

	expectedKeys := []string{"Vstring", "Vunique"}

	sort.Strings(md.Keys)
	if !reflect.DeepEqual(md.Keys, expectedKeys) {
		t.Fatalf("bad keys: %#v", md.Keys)
	}

	expectedUnused := []string{}
	if !reflect.DeepEqual(md.Unused, expectedUnused) {
		t.Fatalf("bad unused: %#v", md.Unused)
	}
}

func TestNonPtrValue(t *testing.T) {
	t.Parallel()

	err := Decode(map[string]interface{}{}, Basic{})
	if err == nil {
		t.Fatal("error should exist")
	}

	if err.Error() != "result must be a pointer" {
		t.Errorf("got unexpected error: %s", err)
	}
}

func TestTagged(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"foo": "bar",
		"bar": "value",
	}

	var result Tagged
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if result.Value != "bar" {
		t.Errorf("value should be 'bar', got: %#v", result.Value)
	}

	if result.Extra != "value" {
		t.Errorf("extra should be 'value', got: %#v", result.Extra)
	}
}

func testSliceInput(t *testing.T, input map[string]interface{}, expected *Slice) {
	var result Slice
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("got error: %s", err)
	}

	if result.Vfoo != expected.Vfoo {
		t.Errorf("Vfoo expected '%s', got '%s'", expected.Vfoo, result.Vfoo)
	}

	if result.Vbar == nil {
		t.Fatalf("Vbar a slice, got '%#v'", result.Vbar)
	}

	if len(result.Vbar) != len(expected.Vbar) {
		t.Errorf("Vbar length should be %d, got %d", len(expected.Vbar), len(result.Vbar))
	}

	for i, v := range result.Vbar {
		if v != expected.Vbar[i] {
			t.Errorf(
				"Vbar[%d] should be '%#v', got '%#v'",
				i, expected.Vbar[i], v)
		}
	}
}

func TestDecodePath(t *testing.T) {
	var document string = `{
    "userContext": {
        "cobrandId": 10000004,
        "channelId": -1,
        "locale": "en_US",
        "tncVersion": 2,
        "applicationId": "17CBE222A42161A3FF450E47CF4C1A00",
        "cobrandConversationCredentials": {
            "sessionToken": "06142010_1:b8d011fefbab8bf1753391b074ffedf9578612d676ed2b7f073b5785b"
        },
		"preferenceInfo": {
            "currencyCode": "USD",
            "timeZone": "PST",
            "dateFormat": "MM/dd/yyyy",
            "currencyNotationType": {
                "currencyNotationType": "SYMBOL"
            },
            "numberFormat": {
                "decimalSeparator": ".",
                "groupingSeparator": ",",
                "groupPattern": "###,##0.##"
            }
        }
    },
    "loginName": "sptest1",
    "userId": 10483860,
    "userType":
        {
        "userTypeId": 1,
        "userTypeName": "normal_user"
        }
}`

	type UserType struct {
		UserTypeId   int
		UserTypeName string
	}

	type NumberFormat struct {
		DecimalSeparator  string `jpath:"userContext.preferenceInfo.numberFormat.decimalSeparator"`
		GroupingSeparator string `jpath:"userContext.preferenceInfo.numberFormat.groupingSeparator"`
		GroupPattern      string `jpath:"userContext.preferenceInfo.numberFormat.groupPattern"`
	}

	type User struct {
		Session   string   `jpath:"userContext.cobrandConversationCredentials.sessionToken"`
		CobrandId int      `jpath:"userContext.cobrandId"`
		UserType  UserType `jpath:"userType"`
		LoginName string   `jpath:"loginName"`
		*NumberFormat
	}

	docScript := []byte(document)
	docMap := map[string]interface{}{}
	err := json.Unmarshal(docScript, &docMap)
	if err != nil {
		t.Fatalf("Unable To Unmarshal Test Document, %s", err)
	}

	user := User{}
	DecodePath(docMap, &user)

	session := "06142010_1:b8d011fefbab8bf1753391b074ffedf9578612d676ed2b7f073b5785b"
	if user.Session != session {
		t.Errorf("user.Session should be '%s', we got '%s'", session, user.Session)
	}

	cobrandId := 10000004
	if user.CobrandId != cobrandId {
		t.Errorf("user.CobrandId should be '%d', we got '%d'", cobrandId, user.CobrandId)
	}

	loginName := "sptest1"
	if user.LoginName != loginName {
		t.Errorf("user.LoginName should be '%s', we got '%s'", loginName, user.LoginName)
	}

	userTypeId := 1
	if user.UserType.UserTypeId != userTypeId {
		t.Errorf("user.UserType.UserTypeId should be '%d', we got '%d'", userTypeId, user.UserType.UserTypeId)
	}

	userTypeName := "normal_user"
	if user.UserType.UserTypeName != userTypeName {
		t.Errorf("user.UserType.UserTypeName should be '%s', we got '%s'", userTypeName, user.UserType.UserTypeName)
	}

	decimalSeparator := "."
	if user.NumberFormat.DecimalSeparator != decimalSeparator {
		t.Errorf("user.NumberFormat.DecimalSeparator should be '%s', we got '%s'", decimalSeparator, user.NumberFormat.DecimalSeparator)
	}

	groupingSeparator := ","
	if user.NumberFormat.GroupingSeparator != groupingSeparator {
		t.Errorf("user.NumberFormat.GroupingSeparator should be '%s', we got '%s'", groupingSeparator, user.NumberFormat.GroupingSeparator)
	}

	groupPattern := "###,##0.##"
	if user.NumberFormat.GroupPattern != groupPattern {
		t.Errorf("user.NumberFormat.GroupPattern should be '%s', we got '%s'", groupPattern, user.NumberFormat.GroupPattern)
	}
}
