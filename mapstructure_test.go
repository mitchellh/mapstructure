package mapstructure

import "testing"

type Basic struct {
	Vstring string
	Vint    int
	Vbool   bool
	Vextra  string
	vsilent bool
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
	Value string `mapstructure:"foo"`
}

func TestBasicTypes(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vstring": "foo",
		"vint":    42,
		"vbool":   true,
		"vsilent": true,
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

	if result.Vbool != true {
		t.Errorf("vbool value should be true: %#v", result.Vbool)
	}

	if result.Vextra != "" {
		t.Errorf("vextra value should be empty: %#v", result.Vextra)
	}

	if result.vsilent != false {
		t.Error("vsilent should not be set, it is unexported")
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

func TestMap(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vfoo": "foo",
		"vother": map[string]interface{}{
			"foo": "foo",
			"bar": "bar",
		},
	}

	var result Map
	err := Decode(input, &result)
	if err != nil {
		t.Errorf("got an error: %s", err)
		t.FailNow()
	}

	if result.Vfoo != "foo" {
		t.Errorf("vfoo value should be 'foo': %#v", result.Vfoo)
	}

	if result.Vother == nil {
		t.Error("vother should not be nil")
		t.FailNow()
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
		t.Errorf("got an err: %s", err)
		t.FailNow()
	}

	if result.Value == nil {
		t.Error("value should not be nil")
		t.FailNow()
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
		t.Errorf("got an err: %s", err.Error())
		t.FailNow()
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
		t.Errorf("got an err: %s", err.Error())
		t.FailNow()
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
		t.Errorf("got unexpected error: %s", err)
		t.FailNow()
	}

	if len(result.Value) != 2 {
		t.Errorf("expected two values, got %d", len(result.Value))
		t.FailNow()
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
		t.Error("error should exist")
		t.FailNow()
	}

	derr, ok := err.(*Error)
	if !ok {
		t.Errorf("error should be kind of Error, instead: %#v", err)
		t.FailNow()
	}

	if derr.Errors[0] != "'Vstring' expected type 'string', got 'int'" {
		t.Errorf("got unexpected error: %s", err)
	}
}

func TestNonPtrValue(t *testing.T) {
	t.Parallel()

	err := Decode(map[string]interface{}{}, Basic{})
	if err == nil {
		t.Error("error should exist")
		t.FailNow()
	}

	if err.Error() != "val must be a pointer" {
		t.Errorf("got unexpected error: %s", err)
	}
}

func TestNonStructValue(t *testing.T) {
	t.Parallel()

	result := 42
	err := Decode(map[string]interface{}{}, &result)
	if err == nil {
		t.Error("error should exist")
		t.FailNow()
	}

	if err.Error() != "val must be an addressable struct" {
		t.Errorf("got unexpected error: %s", err)
	}
}

func TestTagged(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"foo": "bar",
	}

	var result Tagged
	err := Decode(input, &result)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		t.FailNow()
	}

	if result.Value != "bar" {
		t.Errorf("value should be 'bar', got: %#v", result.Value)
	}
}

func testSliceInput(t *testing.T, input map[string]interface{}, expected *Slice) {
	var result Slice
	err := Decode(input, &result)
	if err != nil {
		t.Errorf("got error: %s", err)
		t.FailNow()
	}

	if result.Vfoo != expected.Vfoo {
		t.Errorf("Vfoo expected '%s', got '%s'", expected.Vfoo, result.Vfoo)
	}

	if result.Vbar == nil {
		t.Errorf("Vbar a slice, got '%#v'", result.Vbar)
		t.FailNow()
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
