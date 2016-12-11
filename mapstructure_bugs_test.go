package mapstructure

import "testing"

// GH-1
func TestDecode_NilValue(t *testing.T) {
	input := map[string]interface{}{
		"vfoo":   nil,
		"vother": nil,
	}

	var result Map
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("should not error: %s", err)
	}

	if result.Vfoo != "" {
		t.Fatalf("value should be default: %s", result.Vfoo)
	}

	if result.Vother != nil {
		t.Fatalf("Vother should be nil: %s", result.Vother)
	}
}

// GH-10
func TestDecode_mapInterfaceInterface(t *testing.T) {
	input := map[interface{}]interface{}{
		"vfoo":   nil,
		"vother": nil,
	}

	var result Map
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("should not error: %s", err)
	}

	if result.Vfoo != "" {
		t.Fatalf("value should be default: %s", result.Vfoo)
	}

	if result.Vother != nil {
		t.Fatalf("Vother should be nil: %s", result.Vother)
	}
}

func TestNestedTypePointerWithDefaults(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vfoo": "foo",
		"vbar": map[string]interface{}{
			"vstring": "foo",
			"vint":    42,
			"vbool":   true,
		},
	}

	result:=NestedPointer{
		Vbar: &Basic{
			Vuint: 42,
		},
	}
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

	// this is the error
	if result.Vbar.Vuint != 42 {
		t.Errorf("vuint value should be 42: %#v", result.Vbar.Vint)
	}

}


func TestNestedTypeWithDefaults(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vfoo": "foo",
		"vbar": map[string]interface{}{
			"vstring": "foo",
			"vint":    42,
			"vbool":   true,
		},
	}

	result:=Nested{
		Vbar: Basic{
			Vuint: 42,
		},
	}
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

	// this is the error
	if result.Vbar.Vuint != 42 {
		t.Errorf("vuint value should be 42: %#v", result.Vbar.Vint)
	}

}