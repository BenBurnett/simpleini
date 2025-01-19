package simpleini

import (
	"os"
	"reflect"
	"testing"
)

func TestSnakeToPascal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_case", "SnakeCase"},
		{"snake_case_example", "SnakeCaseExample"},
		{"singleword", "Singleword"},
		{"", ""},
	}

	for _, test := range tests {
		result := snakeToPascal(test.input)
		if result != test.expected {
			t.Errorf("snakeToPascal(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestPascalToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PascalCase", "pascal_case"},
		{"AnotherExample", "another_example"},
		{"Single", "single"},
		{"", ""},
	}

	for _, test := range tests {
		result := pascalToSnake(test.input)
		if result != test.expected {
			t.Errorf("expected %q, got %q", test.expected, result)
		}
	}
}

func TestSubstituteEnvVars(t *testing.T) {
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	tests := []struct {
		input    string
		expected string
	}{
		{"${TEST_ENV_VAR}", "test_value"},
		{"prefix_${TEST_ENV_VAR}_suffix", "prefix_test_value_suffix"},
		{"no_env_var", "no_env_var"},
		{"${NON_EXISTENT_VAR}", ""},
	}

	for _, test := range tests {
		result := substituteEnvVars(test.input)
		if result != test.expected {
			t.Errorf("substituteEnvVars(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestIsValidKey(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"valid_key", true},
		{"validKey123", true},
		{"invalid-key", false},
		{"invalid key", false},
		{"", false},
	}

	for _, test := range tests {
		result := isValidKey(test.input)
		if result != test.expected {
			t.Errorf("isValidKey(%q) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

func TestIsValidSection(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"valid_section", true},
		{"valid.section", true},
		{"validSection123", true},
		{"invalid-section", false},
		{"invalid section", false},
		{"", false},
	}

	for _, test := range tests {
		result := isValidSection(test.input)
		if result != test.expected {
			t.Errorf("isValidSection(%q) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

func TestEnsureValidUTF8(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"valid_utf8", "valid_utf8", false},
		{"\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98", "", true}, // Invalid UTF-8
	}

	for _, test := range tests {
		result, err := ensureValidUTF8(test.input)
		if (err != nil) != test.hasError {
			t.Errorf("ensureValidUTF8(%q) error = %v; expected error = %v", test.input, err != nil, test.hasError)
		}
		if result != test.expected {
			t.Errorf("ensureValidUTF8(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestIsSupportedType(t *testing.T) {
	tests := []struct {
		kind     reflect.Kind
		expected bool
	}{
		{reflect.Int, true},
		{reflect.Int8, true},
		{reflect.Int16, true},
		{reflect.Int32, true},
		{reflect.Int64, true},
		{reflect.Uint, true},
		{reflect.Uint8, true},
		{reflect.Uint16, true},
		{reflect.Uint32, true},
		{reflect.Uint64, true},
		{reflect.Bool, true},
		{reflect.Float32, true},
		{reflect.Float64, true},
		{reflect.String, true},
		{reflect.Struct, false},
		{reflect.Slice, false},
		{reflect.Map, false},
		{reflect.Chan, false},
		{reflect.Func, false},
	}

	for _, test := range tests {
		result := isSupportedType(test.kind)
		if result != test.expected {
			t.Errorf("isSupportedType(%v) = %v; expected %v", test.kind, result, test.expected)
		}
	}
}
