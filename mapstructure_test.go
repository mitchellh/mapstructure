package mapstructure

import "testing"

type Basic struct {
	Vstring string
	Vint int
	Vbool bool
}

func TestBasicTypes(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"vstring": "foo",
		"vint": 42,
		"vbool": true,
	}

	var result Basic
	err := MapToStruct(input, &result)
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
}
