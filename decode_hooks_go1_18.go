//go:build go1.18
// +build go1.18

package mapstructure

import (
	"net/netip"
	"reflect"
)

// StringToNetIPAddrHookFunc returns a DecodeHookFunc that converts
// strings to netip.Addr.
func StringToNetIPAddrHookFunc() DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(netip.Addr{}) {
			return data, nil
		}

		// Convert it by parsing
		return netip.ParseAddr(data.(string))
	}
}

// StringToNetIPAddrPortHookFunc returns a DecodeHookFunc that converts
// strings to netip.AddrPort.
func StringToNetIPAddrPortHookFunc() DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(netip.AddrPort{}) {
			return data, nil
		}

		// Convert it by parsing
		return netip.ParseAddrPort(data.(string))
	}
}
