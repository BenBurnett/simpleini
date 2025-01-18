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
	"sync"
)

// Cache for struct field mappings
var fieldCache sync.Map

var delimiter = "="

// SetDelimiter sets the delimiter for key-value pairs in the INI file.
func SetDelimiter(d string) {
	delimiter = d
}

// getFieldMap returns the field map for the given struct type.
// It uses a cache to avoid recomputing the field map for the same type.
func getFieldMap(t reflect.Type) (map[string]reflect.StructField, error) {
	if fieldMap, found := fieldCache.Load(t); found {
		return fieldMap.(map[string]reflect.StructField), nil
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

		if field.Anonymous {
			if field.Tag.Get("ini") != "" {
				return nil, fmt.Errorf("promoted struct '%s' should not have an ini tag", field.Name)
			}
			embeddedFieldMap, err := getFieldMap(field.Type)
			if err != nil {
				return nil, err
			}
			for k, v := range embeddedFieldMap {
				if _, exists := fieldMap[k]; exists {
					return nil, fmt.Errorf("duplicate tag name '%s' in struct %s", k, t.Name())
				}
				fieldMap[k] = v
			}
		} else {
			fieldMap[tagName] = field
		}
	}
	fieldCache.Store(t, fieldMap)
	return fieldMap, nil
}

// initializePointer initializes a pointer if it is nil and has a value or a default value.
func initializePointer(v reflect.Value, hasValue bool) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() && hasValue {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return v.Elem()
	}
	return v
}

// setFieldValue sets the value of a field based on its type.
func setFieldValue(fieldValue reflect.Value, value string) error {
	// Initialize the pointer if necessary
	fieldValue = initializePointer(fieldValue, value != "")

	// Check if the field is unexported
	if !fieldValue.CanSet() {
		return fmt.Errorf("cannot set unexported field")
	}

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

// setDefaultValues sets the default values for all fields in the struct.
func setDefaultValues(v reflect.Value) error {
	fieldMap, err := getFieldMap(v.Type())
	if err != nil {
		return err
	}

	for _, field := range fieldMap {
		fieldValue := v.FieldByName(field.Name)
		defaultValue := field.Tag.Get("default")
		if defaultValue != "" {
			fieldValue = initializePointer(fieldValue, true)
			if err := setFieldValue(fieldValue, defaultValue); err != nil {
				return err
			}
		}

		// Recursively set default values for nested structs
		if fieldValue.Kind() == reflect.Struct {
			if err := setDefaultValues(fieldValue); err != nil {
				return err
			}
		} else if fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct {
			// Initialize pointer to struct if any field has a default value
			embeddedFieldMap, err := getFieldMap(fieldValue.Type().Elem())
			if err != nil {
				return err
			}
			for _, embeddedField := range embeddedFieldMap {
				if embeddedField.Tag.Get("default") != "" {
					fieldValue = initializePointer(fieldValue, true)
					if err := setDefaultValues(fieldValue); err != nil {
						return err
					}
					break
				}
			}
		}
	}
	return nil
}

// setStructValue sets the value of a field in the struct.
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

	fieldValue := v.FieldByName(field.Name)
	fieldValue = initializePointer(fieldValue, value != "")
	return setFieldValue(fieldValue, value)
}

// setConfigValue sets the value of a field in the config struct.
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
		field = initializePointer(field, true)

		// Check if the field is a struct
		if field.Kind() != reflect.Struct {
			return fmt.Errorf("field for section '%s' is not a struct", section)
		}
		v = field
	}

	return setStructValue(v, key, value)
}

// processMultilineValue processes and sets a multiline value.
func processMultilineValue(config interface{}, section, key, value string, lineNumber int) error {
	value = substituteEnvVars(value)
	if err := setConfigValue(config, section, key, value); err != nil {
		return fmt.Errorf("error at line %d: %w", lineNumber, err)
	}
	return nil
}

// processLine processes a single line from the INI file.
func processLine(line string, config interface{}, currentSection *string, currentKey *string, currentValue *string, inMultiline *bool, lineNumber int) error {
	// Check for multiline continuation
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		*inMultiline = true
		*currentValue += "\n" + strings.TrimSpace(line)
		return nil
	}

	// Process the previous multiline value
	if *inMultiline {
		if err := processMultilineValue(config, *currentSection, *currentKey, *currentValue, lineNumber); err != nil {
			return err
		}
		*inMultiline = false
	}

	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] == ';' || line[0] == '#' {
		return nil
	}

	// Check if the line is a section header
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		section := strings.ToLower(line[1 : len(line)-1])
		if !isValidSection(section) {
			return fmt.Errorf("invalid section name at line %d: %s", lineNumber, section)
		}
		*currentSection = section
	} else {
		// Check if the line is a key-value pair
		if !strings.Contains(line, delimiter) {
			return fmt.Errorf("invalid line format at line %d: %s", lineNumber, line)
		}

		// Split the line into key and value
		keyValue := strings.SplitN(line, delimiter, 2)
		key := strings.ToLower(strings.TrimSpace(keyValue[0]))
		if !isValidKey(key) {
			return fmt.Errorf("invalid key name at line %d: %s", lineNumber, key)
		}
		*currentKey = key
		*currentValue = strings.TrimSpace(keyValue[1])
		*currentValue = substituteEnvVars(*currentValue)

		// Use reflection to set the value in the config struct
		if err := setConfigValue(config, *currentSection, *currentKey, *currentValue); err != nil {
			return fmt.Errorf("error at line %d: %w", lineNumber, err)
		}
	}

	return nil
}

// handleIncludeDirective processes an include directive.
func handleIncludeDirective(line, basePath string, config interface{}, includedFiles map[string]bool, depth int) ([]error, bool) {
	if strings.HasPrefix(line, "!include ") {
		includeFile := strings.TrimSpace(line[len("!include "):])
		if !filepath.IsAbs(includeFile) {
			includeFile = filepath.Join(basePath, includeFile)
		}
		includeErrors := parseFile(includeFile, config, includedFiles, depth)
		return includeErrors, true
	}
	return nil, false
}

// parseReader parses the INI content from an io.Reader with support for include directives.
func parseReader(reader io.Reader, config interface{}, includedFiles map[string]bool, depth int, basePath string) []error {
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

		// Ensure the line is valid UTF-8
		line, err := ensureValidUTF8(line)
		if err != nil {
			errors = append(errors, fmt.Errorf("error at line %d: %w", lineNumber, err))
			continue
		}

		// Handle include directive
		if includeErrors, handled := handleIncludeDirective(line, basePath, config, includedFiles, depth); handled {
			if includeErrors != nil {
				errors = append(errors, includeErrors...)
			}
			continue
		}

		// Process the line
		if err := processLine(line, config, &currentSection, &currentKey, &currentValue, &inMultiline, lineNumber); err != nil {
			errors = append(errors, err)
		}
	}

	// Process any remaining multiline value
	if inMultiline {
		if err := processMultilineValue(config, currentSection, currentKey, currentValue, lineNumber); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// parseFile reads and parses an INI file with support for include directives.
func parseFile(filename string, config interface{}, includedFiles map[string]bool, depth int) []error {
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
	return parseReader(file, config, includedFiles, depth+1, basePath)
}

// Parse parses the INI file content from an io.Reader and populates the config struct.
func Parse(reader io.Reader, config interface{}) []error {
	return parseReader(reader, config, make(map[string]bool), 0, "")
}
