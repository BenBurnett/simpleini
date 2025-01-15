package simpleini

import (
	"bufio"
	"encoding"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// Cache for struct field mappings
var fieldCache = make(map[reflect.Type]map[string]reflect.StructField)

// Parse parses the INI file content from an io.Reader and populates the config struct
func Parse(reader io.Reader, config interface{}) error {
	scanner := bufio.NewScanner(reader)
	var currentSection string

	// Read the file line by line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == ';' || line[0] == '#' {
			continue
		}

		// Check if the line is a section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
		} else {
			// Check if the line is a key-value pair
			if !strings.Contains(line, "=") {
				return fmt.Errorf("invalid line format: %s", line)
			}

			// Split the line into key and value
			keyValue := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(keyValue[0])
			value := strings.TrimSpace(keyValue[1])

			// Use reflection to set the value in the config struct
			if err := setConfigValue(config, currentSection, key, value); err != nil {
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
		// Find the field by tag or converted name
		field := v.FieldByNameFunc(func(name string) bool {
			field, ok := v.Type().FieldByName(name)
			return ok && (field.Tag.Get("ini") == part || snakeToPascal(part) == name)
		})

		// If the field is not found, return an error
		if !field.IsValid() {
			return fmt.Errorf("no matching field found for section '%s'", section)
		}

		// If the field is a pointer, initialize it
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}

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
	// Get the field map for the struct type
	fieldMap, found := fieldCache[v.Type()]

	// If the field map is not found, create it
	if !found {
		fieldMap = make(map[string]reflect.StructField)
		for i := 0; i < v.Type().NumField(); i++ {
			field := v.Type().Field(i)
			fieldMap[field.Tag.Get("ini")] = field
			fieldMap[snakeToPascal(field.Name)] = field
		}
		fieldCache[v.Type()] = fieldMap
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
	// If the field is a pointer, initialize it
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		fieldValue = fieldValue.Elem()
	}

	// Check if the field implements encoding.TextUnmarshaler, and if so, use it
	if fieldValue.CanAddr() {
		addr := fieldValue.Addr()
		if addr.CanInterface() && addr.Type().Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
			return addr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
		}
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
