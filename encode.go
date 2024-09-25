package jprop

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := marshalValue(reflect.ValueOf(v), &buf, "")
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

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
					err := marshalValue(fieldValue, buf, fullKey+".")
					if err != nil {
						return err
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
		for _, key := range val.MapKeys() {
			mapValue := val.MapIndex(key)
			strKey, err := valueToString(key)
			if err != nil {
				return err
			}
			fullKey := prefix + strKey
			if mapValue.Kind() == reflect.Struct || mapValue.Kind() == reflect.Map {
				err := marshalValue(mapValue, buf, fullKey+".")
				if err != nil {
					return err
				}
			} else {
				strValue, err := valueToString(mapValue)
				if err != nil {
					return err
				}
				buf.WriteString(fmt.Sprintf("%s=%s\n", fullKey, strValue))
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			elemValue := val.Index(i)
			fullKey := fmt.Sprintf("%s[%d]", prefix[:len(prefix)-1], i)
			if elemValue.Kind() == reflect.Struct || elemValue.Kind() == reflect.Map {
				err := marshalValue(elemValue, buf, fullKey+".")
				if err != nil {
					return err
				}
			} else {
				strValue, err := valueToString(elemValue)
				if err != nil {
					return err
				}
				buf.WriteString(fmt.Sprintf("%s=%s\n", fullKey, strValue))
			}
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

func valueToString(v reflect.Value) (string, error) {
	if !v.IsValid() {
		return "", nil
	}

	if v.CanInterface() {
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
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprint(v.Complex()), nil
	default:
		return "", fmt.Errorf("unsupported type: %s", v.Type())
	}
}
