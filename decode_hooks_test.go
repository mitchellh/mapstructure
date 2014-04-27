package mapstructure

import (
	"reflect"
	"testing"
)

func TestStringToSliceHookFunc(t *testing.T) {
	f := StringToSliceHookFunc(",")

	cases := []struct {
		f, t   reflect.Kind
		data   interface{}
		result interface{}
		err    bool
	}{
		{reflect.Slice, reflect.Slice, 42, 42, false},
		{reflect.String, reflect.String, 42, 42, false},
		{
			reflect.String,
			reflect.Slice,
			"foo,bar,baz",
			[]string{"foo", "bar", "baz"},
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
