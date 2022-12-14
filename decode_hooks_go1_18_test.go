//go:build go1.18
// +build go1.18

package mapstructure

import (
	"net/netip"
	"reflect"
	"testing"
)

func TestStringToNetIPAddrHookFunc(t *testing.T) {
	strValue := reflect.ValueOf("5")
	addrValue := reflect.ValueOf(netip.Addr{})
	cases := []struct {
		f, t   reflect.Value
		result interface{}
		err    bool
	}{
		{reflect.ValueOf("192.0.2.1"), addrValue,
			netip.AddrFrom4([4]byte{0xc0, 0x00, 0x02, 0x01}), false},
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
		{reflect.ValueOf("192.0.2.1:80"), addrPortValue,
			netip.AddrPortFrom(netip.AddrFrom4([4]byte{0xc0, 0x00, 0x02, 0x01}), 80), false},
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
