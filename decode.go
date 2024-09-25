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

// Unmarshal loads data from a .properties format into a struct
func Unmarshal(data []byte, v interface{}) error {
	lines := strings.Split(string(data), "\n")
	val := reflect.ValueOf(v).Elem()
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		err := setValueFromString(val, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// setValueFromString gestisce la deserializzazione
func setValueFromString(v reflect.Value, key, value string) error {
	val := reflect.Indirect(v)
	switch val.Kind() {
	case reflect.Struct:
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldValue := val.Field(i)
			tag := field.Tag.Get("jprop")
			tagOptions := parseTagOptions(tag)
			fieldKey := tagOptions.name
			if fieldKey == "" {
				fieldKey = field.Name
			}
			if strings.HasPrefix(key, fieldKey) {
				subKey := strings.TrimPrefix(key, fieldKey)
				if subKey == "" || subKey[0] == '.' {
					subKey = strings.TrimPrefix(subKey, ".")
					if fieldValue.Kind() == reflect.Struct {
						return setValueFromString(fieldValue, subKey, value)
					} else if fieldValue.Kind() == reflect.Slice {
						// Handle slices
						items := strings.Split(value, ",")
						slice := reflect.MakeSlice(fieldValue.Type(), len(items), len(items))
						for idx, item := range items {
							if err := setBasicValue(slice.Index(idx), strings.TrimSpace(item)); err != nil {
								return err
							}
						}
						fieldValue.Set(slice)
						return nil
					} else if fieldValue.Kind() == reflect.Map {
						// Handle maps
						mapKey := extractMapKey(subKey) // Extract only the key
						if mapKey != "" {
							// Ensure the map is initialized
							if fieldValue.IsNil() {
								fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
							}
							// Set the value in the map
							mapValue := fieldValue.MapIndex(reflect.ValueOf(mapKey))
							if !mapValue.IsValid() {
								// If it doesn't exist, create a new value for the key
								mapValue = reflect.New(fieldValue.Type().Elem()).Elem()
							}
							// Set the value using setBasicValue
							if err := setBasicValue(mapValue, value); err != nil {
								return err
							}
							fieldValue.SetMapIndex(reflect.ValueOf(mapKey), mapValue)
							return nil
						}
					}
					return setValueFromString(fieldValue, subKey, value)
				}
			}
		}
	case reflect.Map:
		// Handle maps
		mapKey := extractMapKey(key)
		if mapKey != "" {
			// Ensure the map is initialized
			if val.IsNil() {
				val.Set(reflect.MakeMap(val.Type()))
			}
			// Create or update the map entry
			mapValue := val.MapIndex(reflect.ValueOf(mapKey))
			if !mapValue.IsValid() {
				// If it doesn't exist, create a new value
				mapValue = reflect.New(val.Type().Elem()).Elem()
			}
			if err := setBasicValue(mapValue, value); err != nil {
				return err
			}
			val.SetMapIndex(reflect.ValueOf(mapKey), mapValue)
			return nil
		}
	default:
		return setBasicValue(val, value)
	}
	return nil
}

// extractMapKey extracts the key of the map from the given string
func extractMapKey(key string) string {
	parts := strings.SplitN(key, ".", 2)
	return parts[0] // Return only the key part
}

// setBasicValue sets the value of a basic field type (string, bool, int, float, etc.)
func setBasicValue(v reflect.Value, value string) error {
	if !v.IsValid() {
		return fmt.Errorf("invalid value provided")
	}
	if !v.CanSet() {
		fmt.Println(v)
		return fmt.Errorf("cannot set value on provided reflect.Value")
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(value)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(boolVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(intVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		v.SetFloat(floatVal)
	default:
		return fmt.Errorf("unsupported type: %s", v.Type())
	}
	return nil
}
