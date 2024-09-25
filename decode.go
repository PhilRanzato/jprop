package jprop

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshaler defines an interface for unmarshaling properties.
type Unmarshaler interface {
	UnmarshalProperties(string) error
}

// Unmarshal loads data from a .properties format into a struct.
func Unmarshal(data []byte, v interface{}) error {
	lines := strings.Split(string(data), "\n")
	val := reflect.ValueOf(v).Elem()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if err := setValueFromString(val, key, value); err != nil {
			return err
		}
	}
	return nil
}

// setValueFromString manages the deserialization process.
func setValueFromString(v reflect.Value, key, value string) error {
	val := reflect.Indirect(v)

	switch val.Kind() {
	case reflect.Struct:
		return setStructValue(val, key, value)
	case reflect.Map:
		return setMapValue(val, key, value)
	default:
		return setBasicValue(val, value)
	}
}

func setStructValue(val reflect.Value, key, value string) error {
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		tag := field.Tag.Get("jprop")
		fieldKey := parseTagOptions(tag).name
		if fieldKey == "" {
			fieldKey = field.Name
		}

		if strings.HasPrefix(key, fieldKey) {
			subKey := strings.TrimPrefix(key, fieldKey)
			if subKey == "" || subKey[0] == '.' {
				subKey = strings.TrimPrefix(subKey, ".")
				if err := handleFieldType(fieldValue, subKey, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func handleFieldType(fieldValue reflect.Value, subKey, value string) error {
	switch fieldValue.Kind() {
	case reflect.Struct:
		return setValueFromString(fieldValue, subKey, value)
	case reflect.Slice:
		return setSliceValue(fieldValue, value)
	case reflect.Map:
		return setMapValue(fieldValue, subKey, value)
	default:
		return setValueFromString(fieldValue, subKey, value)
	}
}

func setSliceValue(fieldValue reflect.Value, value string) error {
	items := strings.Split(value, ",")
	slice := reflect.MakeSlice(fieldValue.Type(), len(items), len(items))
	for idx, item := range items {
		if err := setBasicValue(slice.Index(idx), strings.TrimSpace(item)); err != nil {
			return err
		}
	}
	fieldValue.Set(slice)
	return nil
}

func setMapValue(val reflect.Value, key, value string) error {
	mapKey := extractMapKey(key)

	// Initialize the map if it is nil
	if val.IsNil() {
		val.Set(reflect.MakeMap(val.Type()))
	}

	// Handle empty keys correctly
	if mapKey == "" {
		// Set the empty key value
		mapValue := reflect.New(val.Type().Elem()).Elem()
		if err := setBasicValue(mapValue, value); err != nil {
			return err
		}
		val.SetMapIndex(reflect.ValueOf(mapKey), mapValue)
		return nil
	}

	// Retrieve the existing value for the given map key
	mapValue := val.MapIndex(reflect.ValueOf(mapKey))

	// If it doesn't exist, create a new value
	if !mapValue.IsValid() {
		mapValue = reflect.New(val.Type().Elem()).Elem()
	}

	// Set the value in the map
	if err := setBasicValue(mapValue, value); err != nil {
		return err
	}
	val.SetMapIndex(reflect.ValueOf(mapKey), mapValue)

	return nil
}

// extractMapKey extracts the key of the map from the given string.
func extractMapKey(key string) string {
	parts := strings.SplitN(key, ".", 2)
	return parts[0] // Return only the key part
}

// setBasicValue sets the value of a basic field type (string, bool, int, float, etc.).
func setBasicValue(v reflect.Value, value string) error {
	if !v.IsValid() || !v.CanSet() {
		return fmt.Errorf("invalid or unassignable value provided: %s", v)
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
