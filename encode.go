package jprop

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Marshaler interface {
	MarshalProperties() (string, error)
}

// Marshal serializes a struct into a .properties file format
func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := marshalValue(reflect.ValueOf(v), &buf, "")
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// marshalValue handles serialization of structs, maps, and slices
func marshalValue(val reflect.Value, buf *bytes.Buffer, prefix string) error {
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Struct:
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldValue := val.Field(i)
			tag := field.Tag.Get("jprop")
			if tag == "-" {
				continue
			}
			tagOptions := parseTagOptions(tag)
			key := tagOptions.name
			if key == "" {
				key = field.Name
			}
			fullKey := prefix + key
			if tagOptions.omitEmpty && isEmptyValue(fieldValue) {
				continue
			}
			if fieldValue.CanInterface() {
				if fieldValue.Kind() == reflect.Struct {
					// Handle nested structs
					err := marshalValue(fieldValue, buf, fullKey+".")
					if err != nil {
						return err
					}
				} else if fieldValue.Kind() == reflect.Slice {
					// Handle slices: serialize on a single line separated by commas
					elements := make([]string, fieldValue.Len())
					for i := 0; i < fieldValue.Len(); i++ {
						elemValue := fieldValue.Index(i)
						strValue, err := valueToString(elemValue)
						if err != nil {
							return err
						}
						elements[i] = strValue
					}
					// Write the slice as a comma-separated string
					buf.WriteString(fmt.Sprintf("%s=%s\n", fullKey, strings.Join(elements, ",")))
				} else if fieldValue.Kind() == reflect.Map {
					// Handle maps
					for _, key := range fieldValue.MapKeys() {
						mapValue := fieldValue.MapIndex(key)
						strValue, err := valueToString(mapValue)
						if err != nil {
							return err
						}
						buf.WriteString(fmt.Sprintf("%s.%s=%s\n", fullKey, key, strValue))
					}
				} else {
					strValue, err := valueToString(fieldValue)
					if err != nil {
						return err
					}
					buf.WriteString(fmt.Sprintf("%s=%s\n", fullKey, strValue))
				}
			}
		}
	case reflect.Map:
		// Handling maps
		for _, key := range val.MapKeys() {
			mapValue := val.MapIndex(key)
			strValue, err := valueToString(mapValue)
			if err != nil {
				return err
			}
			buf.WriteString(fmt.Sprintf("%s=%s\n", key, strValue))
		}
	default:
		strValue, err := valueToString(val)
		if err != nil {
			return err
		}
		buf.WriteString(fmt.Sprintf("%s=%s\n", prefix[:len(prefix)-1], strValue))
	}
	return nil
}

// valueToString converts values into strings
func valueToString(v reflect.Value) (string, error) {
	if !v.IsValid() {
		return "", nil
	}

	if v.CanInterface() {
		// Check if it implements the Marshaler interface
		if m, ok := v.Interface().(Marshaler); ok {
			return m.MarshalProperties()
		}
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("unsupported type: %s", v.Type())
	}
}
