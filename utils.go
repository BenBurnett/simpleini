package simpleini

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// snakeToPascal converts a snake_case string to PascalCase.
func snakeToPascal(s string) string {
	var result strings.Builder
	upperNext := true
	for _, r := range s {
		if r == '_' {
			upperNext = true
			continue
		}

		if upperNext {
			result.WriteRune(unicode.ToUpper(r))
			upperNext = false
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// pascalToSnake converts a PascalCase string to snake_case.
func pascalToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// substituteEnvVars replaces placeholders in the value with environment variable values.
func substituteEnvVars(value string) string {
	return os.Expand(value, func(key string) string {
		return os.Getenv(key)
	})
}

// isValidKey checks if the key contains only valid characters and is not empty.
func isValidKey(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

// isValidSection checks if the section contains only valid characters and is not empty.
func isValidSection(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

// ensureValidUTF8 checks if the input string is valid UTF-8.
func ensureValidUTF8(input string) (string, error) {
	if !utf8.ValidString(input) {
		return "", fmt.Errorf("invalid UTF-8 encoding")
	}
	return input, nil
}

// isSupportedType checks if the given kind is a supported type.
func isSupportedType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Bool, reflect.Float32, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}
