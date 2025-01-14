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

// ParseINI parses the INI file content from an io.Reader and populates the config struct
func ParseINI(reader io.Reader, config interface{}) error {
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
				return fmt.Errorf("invalid line: %s", line)
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
		return err
	}

	return nil
}

func setConfigValue(config interface{}, section, key, value string) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("config must be a pointer to a struct")
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if section == "" && field.Tag.Get("ini") == key {
			return setFieldValue(v.Field(i), value)
		} else if field.Tag.Get("ini") == section {
			fieldValue := v.Field(i)
			if fieldValue.Kind() == reflect.Struct {
				return setStructValue(fieldValue, key, value)
			}
		} else if strings.HasPrefix(section, field.Tag.Get("ini")+".") {
			subSection := strings.TrimPrefix(section, field.Tag.Get("ini")+".")
			fieldValue := v.Field(i)
			if fieldValue.Kind() == reflect.Struct {
				return setConfigValue(fieldValue.Addr().Interface(), subSection, key, value)
			}
		}
	}
	return fmt.Errorf("no matching field found for section: %s, key: %s", section, key)
}

func setStructValue(v reflect.Value, key, value string) error {
	t := v.Type()

	numFields := t.NumField()
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		if field.Tag.Get("ini") == key {
			return setFieldValue(v.Field(i), value)
		}
	}
	return fmt.Errorf("no matching field found for key: %s", key)
}

func setFieldValue(fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.Int:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int value: %s", value)
		}
		fieldValue.SetInt(int64(intValue))
	case reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		fieldValue.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value: %s", value)
		}
		fieldValue.SetBool(boolValue)
	case reflect.String:
		fieldValue.SetString(value)
	default:
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}
	return nil
}
