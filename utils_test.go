package simpleini

import (
	"os"
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
