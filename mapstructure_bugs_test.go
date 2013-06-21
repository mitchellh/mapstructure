package mapstructure

import "testing"

// GH-1
func TestDecode_NilValue(t *testing.T) {
	input := map[string]interface{}{
		"vstring": nil,
	}

	var result Basic
	err := Decode(input, &result)
	if err != nil {
		t.Fatalf("should not error: %s", err)
	}

	if result.Vstring != "" {
		t.Fatalf("value should be default: %s", result.Vstring)
	}
}

