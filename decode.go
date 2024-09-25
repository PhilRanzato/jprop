package jprop

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Unmarshaler interface {
	UnmarshalProperties(string) error
}

func Unmarshal(data []byte, v interface{}) error {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		err := setValueFromString(v, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func setValueFromString(v interface{}, key, value string) error {
	structValue := reflect.ValueOf(v).Elem()
	structType := structValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		tag := field.Tag.Get("jprop")
		tagOptions := parseTagOptions(tag)
		propName := tagOptions.name
		if propName == "" {
			propName = field.Name
		}

		// Verifica se la chiave corrisponde al campo corrente
		if strings.HasPrefix(key, propName) {
			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(value)
			case reflect.Bool:
				boolValue, err := strconv.ParseBool(value)
				if err != nil {
					return fmt.Errorf("invalid boolean value for key %s: %s", key, value)
				}
				fieldValue.SetBool(boolValue)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				intValue, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid integer value for key %s: %s", key, value)
				}
				fieldValue.SetInt(intValue)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				uintValue, err := strconv.ParseUint(value, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid unsigned integer value for key %s: %s", key, value)
				}
				fieldValue.SetUint(uintValue)
			case reflect.Float32, reflect.Float64:
				floatValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return fmt.Errorf("invalid float value for key %s: %s", key, value)
				}
				fieldValue.SetFloat(floatValue)
			case reflect.Slice:
				elemType := fieldValue.Type().Elem()
				values := strings.Split(value, ",")
				slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))
				for i, v := range values {
					elemValue := reflect.New(elemType).Elem()
					if err := setBasicTypeFromString(v, elemValue); err != nil {
						return err
					}
					slice.Index(i).Set(elemValue)
				}
				fieldValue.Set(slice)
			case reflect.Map:
				if fieldValue.Type().Key().Kind() != reflect.String {
					return fmt.Errorf("unsupported map key type for key %s: %s", key, fieldValue.Type().Key())
				}
				if fieldValue.IsNil() {
					fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
				}

				mapKeyValue := strings.SplitN(key, ".", 2)
				if len(mapKeyValue) < 2 {
					return fmt.Errorf("invalid map key format for key %s", key)
				}
				mapKey := mapKeyValue[1]

				mapKeyElem := reflect.New(fieldValue.Type().Key()).Elem()
				if err := setBasicTypeFromString(mapKey, mapKeyElem); err != nil {
					return err
				}

				mapValueElem := reflect.New(fieldValue.Type().Elem()).Elem()
				if err := setBasicTypeFromString(value, mapValueElem); err != nil {
					return err
				}
				fieldValue.SetMapIndex(mapKeyElem, mapValueElem)
			default:
				return fmt.Errorf("unsupported field type: %s", fieldValue.Type())
			}
		}
	}

	return nil
}

func setBasicTypeFromString(value string, v reflect.Value) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(value)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(boolValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		v.SetFloat(floatValue)
	default:
		return fmt.Errorf("unsupported type: %s", v.Type())
	}
	return nil
}
