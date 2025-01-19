package simpleini

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
)

// Write writes the config struct to the provided io.Writer in INI format.
func Write(w io.Writer, config interface{}) error {
	fieldCache = sync.Map{} // Clear the field cache

	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.Type().Elem().Kind() != reflect.Struct {
		return errors.New("configuration must be a pointer to a struct")
	}
	v = v.Elem()

	return writeStruct(w, v, "")
}

func writeStruct(w io.Writer, v reflect.Value, section string) error {
	return writeStructHelper(w, v, section, false)
}

func writeStructAsComments(w io.Writer, v reflect.Value, section string) error {
	return writeStructHelper(w, v, section, true)
}

func writeStructHelper(w io.Writer, v reflect.Value, section string, asComments bool) error {
	if err := writeFields(w, v, section, asComments); err != nil {
		return err
	}

	return writeNestedStructs(w, v, section, asComments)
}

func writeFields(w io.Writer, v reflect.Value, section string, asComments bool) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		tagName := field.Tag.Get("ini")
		if tagName == "" {
			tagName = pascalToSnake(field.Name)
		}
		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			if err := writeFields(w, fieldValue, section, asComments); err != nil {
				return err
			}
			continue
		}
		if err := writeField(w, fieldValue, tagName, section, asComments); err != nil {
			return err
		}
	}

	return nil
}

func writeField(w io.Writer, fieldValue reflect.Value, tagName, section string, asComments bool) error {
	if fieldValue.Kind() == reflect.Struct || (fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct) {
		return nil
	}

	if (fieldValue.Kind() == reflect.Ptr && !isSupportedType(fieldValue.Type().Elem().Kind())) || (fieldValue.Kind() != reflect.Ptr && !isSupportedType(fieldValue.Kind())) {
		return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
	}

	if section != "" {
		tagName = strings.TrimPrefix(tagName, section+".")
	}

	var value string
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			value = ""
		} else {
			value = fmt.Sprintf("%v", fieldValue.Elem().Interface())
		}
	} else {
		value = fmt.Sprintf("%v", fieldValue.Interface())
	}

	if asComments {
		_, err := fmt.Fprintf(w, "; %s %s\n", tagName, delimiter)
		return err
	}
	_, err := fmt.Fprintf(w, "%s %s %s\n", tagName, delimiter, value)
	return err
}

func writeNestedStructs(w io.Writer, v reflect.Value, section string, asComments bool) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		tagName := field.Tag.Get("ini")
		if tagName == "" {
			tagName = pascalToSnake(field.Name)
		}
		if fieldValue.Kind() == reflect.Struct && !field.Anonymous {
			newSection := buildSectionName(section, tagName)
			if err := writeSectionHeader(w, newSection, asComments); err != nil {
				return err
			}
			if err := writeStructHelper(w, fieldValue, newSection, asComments); err != nil {
				return err
			}
		} else if fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct {
			newSection := buildSectionName(section, tagName)

			if fieldValue.IsNil() {
				if err := writeSectionHeader(w, newSection, true); err != nil {
					return err
				}
				if err := writeStructAsComments(w, reflect.New(field.Type.Elem()).Elem(), newSection); err != nil {
					return err
				}
			} else {
				if err := writeSectionHeader(w, newSection, asComments); err != nil {
					return err
				}
				if err := writeStructHelper(w, fieldValue.Elem(), newSection, asComments); err != nil {
					return err
				}
			}
		} else if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			if err := writeNestedStructs(w, fieldValue, section, asComments); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildSectionName(section, tagName string) string {
	if section == "" {
		return tagName
	}
	return section + "." + tagName
}

func writeSectionHeader(w io.Writer, section string, asComments bool) error {
	if asComments {
		_, err := fmt.Fprintf(w, "\n; [%s]\n", section)
		return err
	}
	_, err := fmt.Fprintf(w, "\n[%s]\n", section)
	return err
}
