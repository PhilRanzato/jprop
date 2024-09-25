package jprop

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

type Marshaler interface {
	MarshalProperties() (string, error)
}

// Marshal serializza una struct in un formato di file .properties
func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := marshalValue(reflect.ValueOf(v), &buf, "")
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// marshalValue gestisce la serializzazione delle struct, mappe e slice
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
					// Gestisci struct nidificate
					err := marshalValue(fieldValue, buf, fullKey+".")
					if err != nil {
						return err
					}
				} else if fieldValue.Kind() == reflect.Slice {
					// Gestisci slice
					for i := 0; i < fieldValue.Len(); i++ {
						elemValue := fieldValue.Index(i)
						strValue, err := valueToString(elemValue)
						if err != nil {
							return err
						}
						buf.WriteString(fmt.Sprintf("%s[%d]=%s\n", fullKey, i, strValue))
					}
				} else if fieldValue.Kind() == reflect.Map {
					// Gestisci mappe
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
		// Gestione delle mappe
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

// valueToString converte i valori in stringhe
func valueToString(v reflect.Value) (string, error) {
	if !v.IsValid() {
		return "", nil
	}

	if v.CanInterface() {
		// Verifica se implementa l'interfaccia Marshaler
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
