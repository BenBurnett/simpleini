package simpleini

import (
	"bufio"
	"encoding"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// getFieldMap returns the field map for the given struct type
func getFieldMap(t reflect.Type) (map[string]reflect.StructField, error) {
	if fieldMap, found := fieldCache[t]; found {
		return fieldMap, nil
	}

	fieldMap := make(map[string]reflect.StructField)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tagName := field.Tag.Get("ini")
		if tagName == "" {
			tagName = snakeToPascal(field.Name)
		}

		if _, exists := fieldMap[tagName]; exists {
			return nil, fmt.Errorf("duplicate tag name '%s' in struct %s", tagName, t.Name())
		}
		fieldMap[tagName] = field
	}
	fieldCache[t] = fieldMap
	return fieldMap, nil
}

// Cache for struct field mappings
var fieldCache = make(map[reflect.Type]map[string]reflect.StructField)

// substituteEnvVars replaces placeholders with environment variable values
func substituteEnvVars(value string) string {
	return os.Expand(value, func(key string) string {
		return os.Getenv(key)
	})
}

// parseFile reads and parses an INI file with support for include directives
func parseFile(filename string, config interface{}, delimiter string, includedFiles map[string]bool, depth int) []error {
	if depth > 10 {
		return []error{fmt.Errorf("maximum include depth exceeded")}
	}

	if includedFiles[filename] {
		return []error{fmt.Errorf("circular include detected: %s", filename)}
	}
	includedFiles[filename] = true

	file, err := os.Open(filename)
	if err != nil {
		return []error{fmt.Errorf("failed to open file: %w", err)}
	}
	defer file.Close()

	basePath := filepath.Dir(filename)
	return parseReader(file, config, delimiter, includedFiles, depth+1, basePath)
}

// parseReader parses the INI content from an io.Reader with support for include directives
func parseReader(reader io.Reader, config interface{}, delimiter string, includedFiles map[string]bool, depth int, basePath string) []error {
	var errors []error

	// Set default values for all fields
	if err := setDefaultValues(reflect.ValueOf(config).Elem()); err != nil {
		errors = append(errors, err)
	}

	scanner := bufio.NewScanner(reader)
	var currentSection, currentKey, currentValue string
	var inMultiline bool
	lineNumber := 0

	// Read the file line by line
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check for include directive
		if strings.HasPrefix(line, "!include ") {
			includeFile := strings.TrimSpace(line[len("!include "):])
			if !filepath.IsAbs(includeFile) {
				includeFile = filepath.Join(basePath, includeFile)
			}
			includeErrors := parseFile(includeFile, config, delimiter, includedFiles, depth)
			if includeErrors != nil {
				errors = append(errors, includeErrors...)
			}
			continue
		}

		// Check for multiline continuation
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			inMultiline = true
			currentValue += "\n" + strings.TrimSpace(line)
			continue
		}

		// Process the previous multiline value
		if inMultiline {
			currentValue = substituteEnvVars(currentValue)
			if err := setConfigValue(config, currentSection, currentKey, currentValue); err != nil {
				errors = append(errors, fmt.Errorf("error at line %d: %w", lineNumber, err))
			}
			inMultiline = false
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == ';' || line[0] == '#' {
			continue
		}

		// Check if the line is a section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(line[1 : len(line)-1])
		} else {
			// Check if the line is a key-value pair
			if !strings.Contains(line, delimiter) {
				errors = append(errors, fmt.Errorf("invalid line format at line %d: %s", lineNumber, line))
				continue
			}

			// Split the line into key and value
			keyValue := strings.SplitN(line, delimiter, 2)
			currentKey = strings.ToLower(strings.TrimSpace(keyValue[0]))
			currentValue = strings.TrimSpace(keyValue[1])
			currentValue = substituteEnvVars(currentValue)

			// Use reflection to set the value in the config struct
			if err := setConfigValue(config, currentSection, currentKey, currentValue); err != nil {
				errors = append(errors, fmt.Errorf("error at line %d: %w", lineNumber, err))
			}
		}
	}

	// Process any remaining multiline value
	if inMultiline {
		currentValue = substituteEnvVars(currentValue)
		if err := setConfigValue(config, currentSection, currentKey, currentValue); err != nil {
			errors = append(errors, fmt.Errorf("error at line %d: %w", lineNumber, err))
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// Parse parses the INI file content from an io.Reader and populates the config struct
func Parse(reader io.Reader, config interface{}) []error {
	return parseReader(reader, config, "=", make(map[string]bool), 0, "")
}

// ParseWithDelimiter parses the INI file content from an io.Reader with a custom delimiter and populates the config struct
func ParseWithDelimiter(reader io.Reader, config interface{}, delimiter string) []error {
	return parseReader(reader, config, delimiter, make(map[string]bool), 0, "")
}

// initializePointer initializes a pointer if it is nil
func initializePointer(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return v.Elem()
	}
	return v
}

// setDefaultValues sets the default values for all fields in the struct
func setDefaultValues(v reflect.Value) error {
	fieldMap, err := getFieldMap(v.Type())
	if err != nil {
		return err
	}

	for _, field := range fieldMap {
		defaultValue := field.Tag.Get("default")
		if defaultValue != "" {
			fieldValue := initializePointer(v.FieldByName(field.Name))
			if err := setFieldValue(fieldValue, defaultValue); err != nil {
				return err
			}
		}

		// Recursively set default values for nested structs
		fieldValue := initializePointer(v.FieldByName(field.Name))
		if fieldValue.Kind() == reflect.Struct {
			if err := setDefaultValues(fieldValue); err != nil {
				return err
			}
		}
	}
	return nil
}

// setConfigValue sets the value of a field in the config struct
func setConfigValue(config interface{}, section, key, value string) error {
	// Check if the config is a pointer to a struct
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("configuration must be a pointer to a struct")
	}
	v = v.Elem()

	// If no section is specified, set the value in the root struct
	if section == "" {
		return setStructValue(v, key, value)
	}

	// Traverse the struct fields to find the section
	sectionParts := strings.Split(section, ".")
	for _, part := range sectionParts {
		part = strings.ToLower(part)
		// Find the field by tag or converted name
		field := v.FieldByNameFunc(func(name string) bool {
			field, ok := v.Type().FieldByName(name)
			return ok && (strings.EqualFold(field.Tag.Get("ini"), part) || strings.EqualFold(snakeToPascal(part), name))
		})

		// If the field is not found, return an error
		if !field.IsValid() {
			return fmt.Errorf("no matching field found for section '%s'", section)
		}

		// Initialize the pointer if necessary
		field = initializePointer(field)

		// Check if the field is a struct
		if field.Kind() != reflect.Struct {
			return fmt.Errorf("field for section '%s' is not a struct", section)
		}
		v = field
	}

	return setStructValue(v, key, value)
}

// setStructValue sets the value of a field in the struct
func setStructValue(v reflect.Value, key, value string) error {
	fieldMap, err := getFieldMap(v.Type())
	if err != nil {
		return err
	}

	// Find the field by key
	field, ok := fieldMap[key]
	if !ok {
		field, ok = fieldMap[snakeToPascal(key)]
		if !ok {
			return fmt.Errorf("no matching field found for key '%s'", key)
		}
	}

	return setFieldValue(v.FieldByName(field.Name), value)
}

// setFieldValue sets the value of a field
func setFieldValue(fieldValue reflect.Value, value string) error {
	// Initialize the pointer if necessary
	fieldValue = initializePointer(fieldValue)

	// Check if the field implements encoding.TextUnmarshaler, and if so, use it
	if fieldValue.CanAddr() {
		addr := fieldValue.Addr()
		if addr.CanInterface() && addr.Type().Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
			return addr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
		}
	}

	// Handle slices
	if fieldValue.Kind() == reflect.Slice {
		lines := strings.Split(value, "\n")
		slice := reflect.MakeSlice(fieldValue.Type(), len(lines), len(lines))
		for i, line := range lines {
			if err := setFieldValue(slice.Index(i), strings.TrimSpace(line)); err != nil {
				return err
			}
		}
		fieldValue.Set(slice)
		return nil
	}

	// Convert the value to the field type
	var err error
	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var intValue int64
		intValue, err = strconv.ParseInt(value, 10, fieldValue.Type().Bits())
		fieldValue.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var uintValue uint64
		uintValue, err = strconv.ParseUint(value, 10, fieldValue.Type().Bits())
		fieldValue.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		var floatValue float64
		floatValue, err = strconv.ParseFloat(value, fieldValue.Type().Bits())
		fieldValue.SetFloat(floatValue)
	case reflect.Bool:
		var boolValue bool
		boolValue, err = strconv.ParseBool(value)
		fieldValue.SetBool(boolValue)
	case reflect.String:
		fieldValue.SetString(value)
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}

	if err != nil {
		return fmt.Errorf("invalid value for field type %s: %s", fieldValue.Kind(), value)
	}
	return nil
}

// snakeToPascal converts a snake_case string to PascalCase
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
