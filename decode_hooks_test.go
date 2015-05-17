package mapstructure

import (
	"errors"
	"reflect"
	"testing"
)

func TestComposeDecodeHookFunc(t *testing.T) {
	f1 := func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		return data.(string) + "foo", nil
	}

	f2 := func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		return data.(string) + "bar", nil
	}

	f := ComposeDecodeHookFunc(f1, f2)

	result, err := f(reflect.TypeOf(""), reflect.TypeOf([]string{}), "")
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if result.(string) != "foobar" {
		t.Fatalf("bad: %#v", result)
	}
}

func TestComposeDecodeHookFunc_err(t *testing.T) {
	f1 := func(reflect.Type, reflect.Type, interface{}) (interface{}, error) {
		return nil, errors.New("foo")
	}

	f2 := func(reflect.Type, reflect.Type, interface{}) (interface{}, error) {
		panic("NOPE")
	}

	f := ComposeDecodeHookFunc(f1, f2)

	_, err := f(reflect.TypeOf(""), reflect.TypeOf([]string{}), 42)
	if err.Error() != "foo" {
		t.Fatalf("bad: %s", err)
	}
}

func TestComposeDecodeHookFunc_kinds(t *testing.T) {
	var f2From reflect.Type

	f1 := func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		return int(42), nil
	}

	f2 := func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		f2From = f
		return data, nil
	}

	f := ComposeDecodeHookFunc(f1, f2)

	_, err := f(reflect.TypeOf(""), reflect.TypeOf([]string{}), "")
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if f2From.Kind() != reflect.Int {
		t.Fatalf("bad: %#v", f2From)
	}
}

func TestStringToSliceHookFunc(t *testing.T) {
	f := StringToSliceHookFunc(",")

	cases := []struct {
		f, t   reflect.Type
		data   interface{}
		result interface{}
		err    bool
	}{
		{reflect.TypeOf([]string{}), reflect.TypeOf([]string{}), 42, 42, false},
		{reflect.TypeOf(""), reflect.TypeOf(""), 42, 42, false},
		{
			reflect.TypeOf(""),
			reflect.TypeOf([]string{}),
			"foo,bar,baz",
			[]string{"foo", "bar", "baz"},
			false,
		},
		{
			reflect.TypeOf(""),
			reflect.TypeOf([]string{}),
			"",
			[]string{},
			false,
		},
	}

	for i, tc := range cases {
		actual, err := f(tc.f, tc.t, tc.data)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestWeaklyTypedHook(t *testing.T) {
	var f DecodeHookFunc = WeaklyTypedHook

	cases := []struct {
		f, t   reflect.Type
		data   interface{}
		result interface{}
		err    bool
	}{
		// TO STRING
		{
			reflect.TypeOf(true),
			reflect.TypeOf(""),
			false,
			"0",
			false,
		},

		{
			reflect.TypeOf(true),
			reflect.TypeOf(""),
			true,
			"1",
			false,
		},

		{
			reflect.TypeOf(float32(0.)),
			reflect.TypeOf(""),
			float32(7),
			"7",
			false,
		},

		{
			reflect.TypeOf(int(0)),
			reflect.TypeOf(""),
			int(7),
			"7",
			false,
		},

		{
			reflect.TypeOf([]string{}),
			reflect.TypeOf(""),
			[]uint8("foo"),
			"foo",
			false,
		},

		{
			reflect.TypeOf(uint(0)),
			reflect.TypeOf(""),
			uint(7),
			"7",
			false,
		},
	}

	for i, tc := range cases {
		actual, err := f(tc.f, tc.t, tc.data)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}
