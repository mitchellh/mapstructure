package mapstructure

import (
	"encoding/json"
	"errors"
	"math/big"
	"net"
	"net/netip"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestComposeDecodeHookFunc(t *testing.T) {
	f1 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return data.(string) + "foo", nil
	}

	f2 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return data.(string) + "bar", nil
	}

	f := ComposeDecodeHookFunc(f1, f2)

	result, err := DecodeHookExec(
		f, reflect.ValueOf(""), reflect.ValueOf([]byte("")))
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if result.(string) != "foobar" {
		t.Fatalf("bad: %#v", result)
	}
}

func TestComposeDecodeHookFunc_err(t *testing.T) {
	f1 := func(reflect.Kind, reflect.Kind, interface{}) (interface{}, error) {
		return nil, errors.New("foo")
	}

	f2 := func(reflect.Kind, reflect.Kind, interface{}) (interface{}, error) {
		panic("NOPE")
	}

	f := ComposeDecodeHookFunc(f1, f2)

	_, err := DecodeHookExec(
		f, reflect.ValueOf(""), reflect.ValueOf([]byte("")))
	if err.Error() != "foo" {
		t.Fatalf("bad: %s", err)
	}
}

func TestComposeDecodeHookFunc_kinds(t *testing.T) {
	var f2From reflect.Kind

	f1 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return int(42), nil
	}

	f2 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		f2From = f
		return data, nil
	}

	f := ComposeDecodeHookFunc(f1, f2)

	_, err := DecodeHookExec(
		f, reflect.ValueOf(""), reflect.ValueOf([]byte("")))
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if f2From != reflect.Int {
		t.Fatalf("bad: %#v", f2From)
	}
}

func TestOrComposeDecodeHookFunc(t *testing.T) {
	f1 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return data.(string) + "foo", nil
	}

	f2 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return data.(string) + "bar", nil
	}

	f := OrComposeDecodeHookFunc(f1, f2)

	result, err := DecodeHookExec(
		f, reflect.ValueOf(""), reflect.ValueOf([]byte("")))
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if result.(string) != "foo" {
		t.Fatalf("bad: %#v", result)
	}
}

func TestOrComposeDecodeHookFunc_correctValueIsLast(t *testing.T) {
	f1 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return nil, errors.New("f1 error")
	}

	f2 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return nil, errors.New("f2 error")
	}

	f3 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return data.(string) + "bar", nil
	}

	f := OrComposeDecodeHookFunc(f1, f2, f3)

	result, err := DecodeHookExec(
		f, reflect.ValueOf(""), reflect.ValueOf([]byte("")))
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	if result.(string) != "bar" {
		t.Fatalf("bad: %#v", result)
	}
}

func TestOrComposeDecodeHookFunc_err(t *testing.T) {
	f1 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return nil, errors.New("f1 error")
	}

	f2 := func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{},
	) (interface{}, error) {
		return nil, errors.New("f2 error")
	}

	f := OrComposeDecodeHookFunc(f1, f2)

	_, err := DecodeHookExec(
		f, reflect.ValueOf(""), reflect.ValueOf([]byte("")))
	if err == nil {
		t.Fatalf("bad: should return an error")
	}
	if err.Error() != "f1 error\nf2 error\n" {
		t.Fatalf("bad: %s", err)
	}
}

func TestComposeDecodeHookFunc_safe_nofuncs(t *testing.T) {
	f := ComposeDecodeHookFunc()
	type myStruct2 struct {
		MyInt int
	}

	type myStruct1 struct {
		Blah map[string]myStruct2
	}

	src := &myStruct1{Blah: map[string]myStruct2{
		"test": {
			MyInt: 1,
		},
	}}

	dst := &myStruct1{}
	dConf := &DecoderConfig{
		Result:      dst,
		ErrorUnused: true,
		DecodeHook:  f,
	}
	d, err := NewDecoder(dConf)
	if err != nil {
		t.Fatal(err)
	}
	err = d.Decode(src)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStringToSliceHookFunc(t *testing.T) {
	f := StringToSliceHookFunc(",")

	strValue := reflect.ValueOf("42")
	sliceValue := reflect.ValueOf([]string{"42"})
	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{sliceValue, sliceValue, []string{"42"}, false},
		{reflect.ValueOf([]byte("42")), reflect.ValueOf([]byte{}), []byte("42"), false},
		{strValue, strValue, "42", false},
		{
			reflect.ValueOf("foo,bar,baz"),
			sliceValue,
			[]string{"foo", "bar", "baz"},
			false,
		},
		{
			reflect.ValueOf(""),
			sliceValue,
			[]string{},
			false,
		},
	}

	for i, tc := range cases {
		actual, err := DecodeHookExec(f, tc.f, tc.t)
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

func TestStringToTimeDurationHookFunc(t *testing.T) {
	f := StringToTimeDurationHookFunc()

	timeValue := reflect.ValueOf(time.Duration(5))
	strValue := reflect.ValueOf("")
	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{reflect.ValueOf("5s"), timeValue, 5 * time.Second, false},
		{reflect.ValueOf("5"), timeValue, time.Duration(0), true},
		{reflect.ValueOf("5"), strValue, "5", false},
	}

	for i, tc := range cases {
		actual, err := DecodeHookExec(f, tc.f, tc.t)
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

func TestStringToTimeHookFunc(t *testing.T) {
	strValue := reflect.ValueOf("5")
	timeValue := reflect.ValueOf(time.Time{})
	cases := []struct {
		f, t   reflect.Value
		layout string
		result interface{}
		err    bool
	}{
		{
			reflect.ValueOf("2006-01-02T15:04:05Z"), timeValue, time.RFC3339,
			time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC), false,
		},
		{strValue, timeValue, time.RFC3339, time.Time{}, true},
		{strValue, strValue, time.RFC3339, "5", false},
	}

	for i, tc := range cases {
		f := StringToTimeHookFunc(tc.layout)
		actual, err := DecodeHookExec(f, tc.f, tc.t)
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

func TestStringToIPHookFunc(t *testing.T) {
	strValue := reflect.ValueOf("5")
	ipValue := reflect.ValueOf(net.IP{})
	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{
			reflect.ValueOf("1.2.3.4"), ipValue,
			net.IPv4(0x01, 0x02, 0x03, 0x04), false,
		},
		{strValue, ipValue, net.IP{}, true},
		{strValue, strValue, "5", false},
	}

	for i, tc := range cases {
		f := StringToIPHookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
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

func TestStringToIPNetHookFunc(t *testing.T) {
	strValue := reflect.ValueOf("5")
	ipNetValue := reflect.ValueOf(net.IPNet{})
	var nilNet *net.IPNet = nil

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{
			reflect.ValueOf("1.2.3.4/24"), ipNetValue,
			&net.IPNet{
				IP:   net.IP{0x01, 0x02, 0x03, 0x00},
				Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0x00),
			}, false,
		},
		{strValue, ipNetValue, nilNet, true},
		{strValue, strValue, "5", false},
	}

	for i, tc := range cases {
		f := StringToIPNetHookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
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

	strValue := reflect.ValueOf("")
	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		// TO STRING
		{
			reflect.ValueOf(false),
			strValue,
			"0",
			false,
		},

		{
			reflect.ValueOf(true),
			strValue,
			"1",
			false,
		},

		{
			reflect.ValueOf(float32(7)),
			strValue,
			"7",
			false,
		},

		{
			reflect.ValueOf(int(7)),
			strValue,
			"7",
			false,
		},

		{
			reflect.ValueOf([]uint8("foo")),
			strValue,
			"foo",
			false,
		},

		{
			reflect.ValueOf(uint(7)),
			strValue,
			"7",
			false,
		},
	}

	for i, tc := range cases {
		actual, err := DecodeHookExec(f, tc.f, tc.t)
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

func TestStructToMapHookFuncTabled(t *testing.T) {
	var f DecodeHookFunc = RecursiveStructToMapHookFunc()

	type b struct {
		TestKey string
	}

	type a struct {
		Sub b
	}

	testStruct := a{
		Sub: b{
			TestKey: "testval",
		},
	}

	testMap := map[string]interface{}{
		"Sub": map[string]interface{}{
			"TestKey": "testval",
		},
	}

	cases := []struct {
		name     string
		receiver interface{}
		input    interface{}
		expected interface{}
		err      bool
	}{
		{
			"map receiver",
			func() interface{} {
				var res map[string]interface{}
				return &res
			}(),
			testStruct,
			&testMap,
			false,
		},
		{
			"interface receiver",
			func() interface{} {
				var res interface{}
				return &res
			}(),
			testStruct,
			func() interface{} {
				var exp interface{} = testMap
				return &exp
			}(),
			false,
		},
		{
			"slice receiver errors",
			func() interface{} {
				var res []string
				return &res
			}(),
			testStruct,
			new([]string),
			true,
		},
		{
			"slice to slice - no change",
			func() interface{} {
				var res []string
				return &res
			}(),
			[]string{"a", "b"},
			&[]string{"a", "b"},
			false,
		},
		{
			"string to string - no change",
			func() interface{} {
				var res string
				return &res
			}(),
			"test",
			func() *string {
				s := "test"
				return &s
			}(),
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &DecoderConfig{
				DecodeHook: f,
				Result:     tc.receiver,
			}

			d, err := NewDecoder(cfg)
			if err != nil {
				t.Fatalf("unexpected err %#v", err)
			}

			err = d.Decode(tc.input)
			if tc.err != (err != nil) {
				t.Fatalf("expected err %#v", err)
			}

			if !reflect.DeepEqual(tc.expected, tc.receiver) {
				t.Fatalf("expected %#v, got %#v",
					tc.expected, tc.receiver)
			}
		})
	}
}

func TestTextUnmarshallerHookFunc(t *testing.T) {
	type MyString string

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{reflect.ValueOf("42"), reflect.ValueOf(big.Int{}), big.NewInt(42), false},
		{reflect.ValueOf("invalid"), reflect.ValueOf(big.Int{}), nil, true},
		{reflect.ValueOf("5"), reflect.ValueOf("5"), "5", false},
		{reflect.ValueOf(json.Number("42")), reflect.ValueOf(big.Int{}), big.NewInt(42), false},
		{reflect.ValueOf(MyString("42")), reflect.ValueOf(big.Int{}), big.NewInt(42), false},
	}
	for i, tc := range cases {
		f := TextUnmarshallerHookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
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

func TestStringToNetIPAddrHookFunc(t *testing.T) {
	strValue := reflect.ValueOf("5")
	addrValue := reflect.ValueOf(netip.Addr{})
	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{
			reflect.ValueOf("192.0.2.1"), addrValue,
			netip.AddrFrom4([4]byte{0xc0, 0x00, 0x02, 0x01}), false,
		},
		{strValue, addrValue, netip.Addr{}, true},
		{strValue, strValue, "5", false},
	}

	for i, tc := range cases {
		f := StringToNetIPAddrHookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
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

func TestStringToNetIPAddrPortHookFunc(t *testing.T) {
	strValue := reflect.ValueOf("5")
	addrPortValue := reflect.ValueOf(netip.AddrPort{})
	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{
			reflect.ValueOf("192.0.2.1:80"), addrPortValue,
			netip.AddrPortFrom(netip.AddrFrom4([4]byte{0xc0, 0x00, 0x02, 0x01}), 80), false,
		},
		{strValue, addrPortValue, netip.AddrPort{}, true},
		{strValue, strValue, "5", false},
	}

	for i, tc := range cases {
		f := StringToNetIPAddrPortHookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
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

func TestStringToBasicTypeHookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, strValue, "42", false},
		{strValue, reflect.ValueOf(int8(0)), int8(42), false},
		{strValue, reflect.ValueOf(uint8(0)), uint8(42), false},
		{strValue, reflect.ValueOf(int16(0)), int16(42), false},
		{strValue, reflect.ValueOf(uint16(0)), uint16(42), false},
		{strValue, reflect.ValueOf(int32(0)), int32(42), false},
		{strValue, reflect.ValueOf(uint32(0)), uint32(42), false},
		{strValue, reflect.ValueOf(int64(0)), int64(42), false},
		{strValue, reflect.ValueOf(uint64(0)), uint64(42), false},
		{strValue, reflect.ValueOf(int(0)), int(42), false},
		{strValue, reflect.ValueOf(uint(0)), uint(42), false},
		{strValue, reflect.ValueOf(float32(0)), float32(42), false},
		{strValue, reflect.ValueOf(float64(0)), float64(42), false},
		{reflect.ValueOf("true"), reflect.ValueOf(bool(false)), true, false},
		{strValue, reflect.ValueOf(byte(0)), byte(42), false},
		{strValue, reflect.ValueOf(rune(0)), rune(42), false},
		{strValue, reflect.ValueOf(complex64(0)), complex64(42), false},
		{strValue, reflect.ValueOf(complex128(0)), complex128(42), false},
	}

	for i, tc := range cases {
		f := StringToBasicTypeHookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToInt8HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	int8Value := reflect.ValueOf(int8(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, int8Value, int8(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), int8Value, int8(0), true},
		{reflect.ValueOf("42.42"), int8Value, int8(0), true},
		{reflect.ValueOf("-42"), int8Value, int8(-42), false},
		{reflect.ValueOf("0b101010"), int8Value, int8(42), false},
		{reflect.ValueOf("052"), int8Value, int8(42), false},
		{reflect.ValueOf("0o52"), int8Value, int8(42), false},
		{reflect.ValueOf("0x2a"), int8Value, int8(42), false},
		{reflect.ValueOf("0X2A"), int8Value, int8(42), false},
		{reflect.ValueOf("0"), int8Value, int8(0), false},
		{reflect.ValueOf("0.0"), int8Value, int8(0), true},
	}

	for i, tc := range cases {
		f := StringToInt8HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToUint8HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	uint8Value := reflect.ValueOf(uint8(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, uint8Value, uint8(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), uint8Value, uint8(0), true},
		{reflect.ValueOf("42.42"), uint8Value, uint8(0), true},
		{reflect.ValueOf("-42"), uint8Value, uint8(0), true},
		{reflect.ValueOf("0b101010"), uint8Value, uint8(42), false},
		{reflect.ValueOf("052"), uint8Value, uint8(42), false},
		{reflect.ValueOf("0o52"), uint8Value, uint8(42), false},
		{reflect.ValueOf("0x2a"), uint8Value, uint8(42), false},
		{reflect.ValueOf("0X2A"), uint8Value, uint8(42), false},
		{reflect.ValueOf("0"), uint8Value, uint8(0), false},
		{reflect.ValueOf("0.0"), uint8Value, uint8(0), true},
	}

	for i, tc := range cases {
		f := StringToUint8HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToInt16HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	int16Value := reflect.ValueOf(int16(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, int16Value, int16(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), int16Value, int16(0), true},
		{reflect.ValueOf("42.42"), int16Value, int16(0), true},
		{reflect.ValueOf("-42"), int16Value, int16(-42), false},
		{reflect.ValueOf("0b101010"), int16Value, int16(42), false},
		{reflect.ValueOf("052"), int16Value, int16(42), false},
		{reflect.ValueOf("0o52"), int16Value, int16(42), false},
		{reflect.ValueOf("0x2a"), int16Value, int16(42), false},
		{reflect.ValueOf("0X2A"), int16Value, int16(42), false},
		{reflect.ValueOf("0"), int16Value, int16(0), false},
		{reflect.ValueOf("0.0"), int16Value, int16(0), true},
	}

	for i, tc := range cases {
		f := StringToInt16HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToUint16HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	uint16Value := reflect.ValueOf(uint16(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, uint16Value, uint16(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), uint16Value, uint16(0), true},
		{reflect.ValueOf("42.42"), uint16Value, uint16(0), true},
		{reflect.ValueOf("-42"), uint16Value, uint16(0), true},
		{reflect.ValueOf("0b101010"), uint16Value, uint16(42), false},
		{reflect.ValueOf("052"), uint16Value, uint16(42), false},
		{reflect.ValueOf("0o52"), uint16Value, uint16(42), false},
		{reflect.ValueOf("0x2a"), uint16Value, uint16(42), false},
		{reflect.ValueOf("0X2A"), uint16Value, uint16(42), false},
		{reflect.ValueOf("0"), uint16Value, uint16(0), false},
		{reflect.ValueOf("0.0"), uint16Value, uint16(0), true},
	}

	for i, tc := range cases {
		f := StringToUint16HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToInt32HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	int32Value := reflect.ValueOf(int32(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, int32Value, int32(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), int32Value, int32(0), true},
		{reflect.ValueOf("42.42"), int32Value, int32(0), true},
		{reflect.ValueOf("-42"), int32Value, int32(-42), false},
		{reflect.ValueOf("0b101010"), int32Value, int32(42), false},
		{reflect.ValueOf("052"), int32Value, int32(42), false},
		{reflect.ValueOf("0o52"), int32Value, int32(42), false},
		{reflect.ValueOf("0x2a"), int32Value, int32(42), false},
		{reflect.ValueOf("0X2A"), int32Value, int32(42), false},
		{reflect.ValueOf("0"), int32Value, int32(0), false},
		{reflect.ValueOf("0.0"), int32Value, int32(0), true},
	}

	for i, tc := range cases {
		f := StringToInt32HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToUint32HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	uint32Value := reflect.ValueOf(uint32(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, uint32Value, uint32(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), uint32Value, uint32(0), true},
		{reflect.ValueOf("42.42"), uint32Value, uint32(0), true},
		{reflect.ValueOf("-42"), uint32Value, uint32(0), true},
		{reflect.ValueOf("0b101010"), uint32Value, uint32(42), false},
		{reflect.ValueOf("052"), uint32Value, uint32(42), false},
		{reflect.ValueOf("0o52"), uint32Value, uint32(42), false},
		{reflect.ValueOf("0x2a"), uint32Value, uint32(42), false},
		{reflect.ValueOf("0X2A"), uint32Value, uint32(42), false},
		{reflect.ValueOf("0"), uint32Value, uint32(0), false},
		{reflect.ValueOf("0.0"), uint32Value, uint32(0), true},
	}

	for i, tc := range cases {
		f := StringToUint32HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToInt64HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	int64Value := reflect.ValueOf(int64(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, int64Value, int64(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), int64Value, int64(0), true},
		{reflect.ValueOf("42.42"), int64Value, int64(0), true},
		{reflect.ValueOf("-42"), int64Value, int64(-42), false},
		{reflect.ValueOf("0b101010"), int64Value, int64(42), false},
		{reflect.ValueOf("052"), int64Value, int64(42), false},
		{reflect.ValueOf("0o52"), int64Value, int64(42), false},
		{reflect.ValueOf("0x2a"), int64Value, int64(42), false},
		{reflect.ValueOf("0X2A"), int64Value, int64(42), false},
		{reflect.ValueOf("0"), int64Value, int64(0), false},
		{reflect.ValueOf("0.0"), int64Value, int64(0), true},
	}

	for i, tc := range cases {
		f := StringToInt64HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToUint64HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	uint64Value := reflect.ValueOf(uint64(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, uint64Value, uint64(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), uint64Value, uint64(0), true},
		{reflect.ValueOf("42.42"), uint64Value, uint64(0), true},
		{reflect.ValueOf("-42"), uint64Value, uint64(0), true},
		{reflect.ValueOf("0b101010"), uint64Value, uint64(42), false},
		{reflect.ValueOf("052"), uint64Value, uint64(42), false},
		{reflect.ValueOf("0o52"), uint64Value, uint64(42), false},
		{reflect.ValueOf("0x2a"), uint64Value, uint64(42), false},
		{reflect.ValueOf("0X2A"), uint64Value, uint64(42), false},
		{reflect.ValueOf("0"), uint64Value, uint64(0), false},
		{reflect.ValueOf("0.0"), uint64Value, uint64(0), true},
	}

	for i, tc := range cases {
		f := StringToUint64HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToIntHookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	intValue := reflect.ValueOf(int(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, intValue, int(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), intValue, int(0), true},
		{reflect.ValueOf("42.42"), intValue, int(0), true},
		{reflect.ValueOf("-42"), intValue, int(-42), false},
		{reflect.ValueOf("0b101010"), intValue, int(42), false},
		{reflect.ValueOf("052"), intValue, int(42), false},
		{reflect.ValueOf("0o52"), intValue, int(42), false},
		{reflect.ValueOf("0x2a"), intValue, int(42), false},
		{reflect.ValueOf("0X2A"), intValue, int(42), false},
		{reflect.ValueOf("0"), intValue, int(0), false},
		{reflect.ValueOf("0.0"), intValue, int(0), true},
	}

	for i, tc := range cases {
		f := StringToIntHookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToUintHookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42")
	uintValue := reflect.ValueOf(uint(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, uintValue, uint(42), false},
		{strValue, strValue, "42", false},
		{reflect.ValueOf(strings.Repeat("42", 42)), uintValue, uint(0), true},
		{reflect.ValueOf("42.42"), uintValue, uint(0), true},
		{reflect.ValueOf("-42"), uintValue, uint(0), true},
		{reflect.ValueOf("0b101010"), uintValue, uint(42), false},
		{reflect.ValueOf("052"), uintValue, uint(42), false},
		{reflect.ValueOf("0o52"), uintValue, uint(42), false},
		{reflect.ValueOf("0x2a"), uintValue, uint(42), false},
		{reflect.ValueOf("0X2A"), uintValue, uint(42), false},
		{reflect.ValueOf("0"), uintValue, uint(0), false},
		{reflect.ValueOf("0.0"), uintValue, uint(0), true},
	}

	for i, tc := range cases {
		f := StringToUintHookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, tc.err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToFloat32HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42.42")
	float32Value := reflect.ValueOf(float32(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, float32Value, float32(42.42), false},
		{strValue, strValue, "42.42", false},
		{reflect.ValueOf(strings.Repeat("42", 420)), float32Value, float32(0), true},
		{reflect.ValueOf("42.42.42"), float32Value, float32(0), true},
		{reflect.ValueOf("-42.42"), float32Value, float32(-42.42), false},
		{reflect.ValueOf("0"), float32Value, float32(0), false},
		{reflect.ValueOf("1e3"), float32Value, float32(1000), false},
		{reflect.ValueOf("1e-3"), float32Value, float32(0.001), false},
	}

	for i, tc := range cases {
		f := StringToFloat32HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToFloat64HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42.42")
	float64Value := reflect.ValueOf(float64(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, float64Value, float64(42.42), false},
		{strValue, strValue, "42.42", false},
		{reflect.ValueOf(strings.Repeat("42", 420)), float64Value, float64(0), true},
		{reflect.ValueOf("42.42.42"), float64Value, float64(0), true},
		{reflect.ValueOf("-42.42"), float64Value, float64(-42.42), false},
		{reflect.ValueOf("0"), float64Value, float64(0), false},
		{reflect.ValueOf("0.0"), float64Value, float64(0), false},
		{reflect.ValueOf("1e3"), float64Value, float64(1000), false},
		{reflect.ValueOf("1e-3"), float64Value, float64(0.001), false},
	}

	for i, tc := range cases {
		f := StringToFloat64HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToComplex64HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42.42+42.42i")
	complex64Value := reflect.ValueOf(complex64(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, complex64Value, complex(float32(42.42), float32(42.42)), false},
		{strValue, strValue, "42.42+42.42i", false},
		{reflect.ValueOf(strings.Repeat("42", 420)), complex64Value, complex(float32(0), 0), true},
		{reflect.ValueOf("42.42.42"), complex64Value, complex(float32(0), 0), true},
		{reflect.ValueOf("-42.42"), complex64Value, complex(float32(-42.42), 0), false},
		{reflect.ValueOf("0"), complex64Value, complex(float32(0), 0), false},
		{reflect.ValueOf("0.0"), complex64Value, complex(float32(0), 0), false},
		{reflect.ValueOf("1e3"), complex64Value, complex(float32(1000), 0), false},
		{reflect.ValueOf("1e-3"), complex64Value, complex(float32(0.001), 0), false},
		{reflect.ValueOf("1e3i"), complex64Value, complex(float32(0), 1000), false},
		{reflect.ValueOf("1e-3i"), complex64Value, complex(float32(0), 0.001), false},
	}

	for i, tc := range cases {
		f := StringToComplex64HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}

func TestStringToComplex128HookFunc(t *testing.T) {
	strValue := reflect.ValueOf("42.42+42.42i")
	complex128Value := reflect.ValueOf(complex128(0))

	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{strValue, complex128Value, complex(42.42, 42.42), false},
		{strValue, strValue, "42.42+42.42i", false},
		{reflect.ValueOf(strings.Repeat("42", 420)), complex128Value, complex(0, 0), true},
		{reflect.ValueOf("42.42.42"), complex128Value, complex(0, 0), true},
		{reflect.ValueOf("-42.42"), complex128Value, complex(-42.42, 0), false},
		{reflect.ValueOf("0"), complex128Value, complex(0, 0), false},
		{reflect.ValueOf("0.0"), complex128Value, complex(0, 0), false},
		{reflect.ValueOf("1e3"), complex128Value, complex(1000, 0), false},
		{reflect.ValueOf("1e-3"), complex128Value, complex(0.001, 0), false},
		{reflect.ValueOf("1e3i"), complex128Value, complex(0, 1000), false},
		{reflect.ValueOf("1e-3i"), complex128Value, complex(0, 0.001), false},
	}

	for i, tc := range cases {
		f := StringToComplex128HookFunc()
		actual, err := DecodeHookExec(f, tc.f, tc.t)
		if tc.err != (err != nil) {
			t.Fatalf("case %d: expected err %#v", i, err)
		}
		if !tc.err && !reflect.DeepEqual(actual, tc.result) {
			t.Fatalf(
				"case %d: expected %#v, got %#v",
				i, tc.result, actual)
		}
	}
}
