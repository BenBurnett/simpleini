package simpleini

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Parse parses the INI file content from an io.Reader and populates the config struct
func Parse(reader io.Reader, config interface{}) error {
	scanner := bufio.NewScanner(reader)
	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
		} else {
			keyValue := strings.SplitN(line, "=", 2)
			if len(keyValue) != 2 {
				return fmt.Errorf("invalid line format: %s", line)
			}
			key := strings.TrimSpace(keyValue[0])
			value := strings.TrimSpace(keyValue[1])

			// Use reflection to set the value in the config struct
			if err := setConfigValue(config, currentSection, key, value); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	return nil
}

func setConfigValue(config interface{}, section, key, value string) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("configuration must be a pointer to a struct")
	}
	v = v.Elem()

	if section == "" {
		return setStructValue(v, key, value)
	}

	sectionParts := strings.Split(section, ".")
	for _, part := range sectionParts {
		field := v.FieldByNameFunc(func(name string) bool {
			if field, ok := v.Type().FieldByName(name); ok {
				return field.Tag.Get("ini") == part
			}
			return false
		})
		if !field.IsValid() {
			return fmt.Errorf("no matching field found for section '%s'", section)
		}
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}
		if field.Kind() != reflect.Struct {
			return fmt.Errorf("field for section '%s' is not a struct", section)
		}
		v = field
	}

	return setStructValue(v, key, value)
}

func setStructValue(v reflect.Value, key, value string) error {
	t := v.Type()

	fieldMap := make(map[string]reflect.StructField)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldMap[field.Tag.Get("ini")] = field
	}

	if field, ok := fieldMap[key]; ok {
		return setFieldValue(v.FieldByName(field.Name), value)
	}
	return fmt.Errorf("no matching field found for key '%s'", key)
}

func setFieldValue(fieldValue reflect.Value, value string) error {
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		fieldValue = fieldValue.Elem()
	}

	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, fieldValue.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
		fieldValue.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, fieldValue.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value: %s", value)
		}
		fieldValue.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, fieldValue.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		fieldValue.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		fieldValue.SetBool(boolValue)
	case reflect.String:
		fieldValue.SetString(value)
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}
	return nil
}
