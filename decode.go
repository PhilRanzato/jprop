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

// Unmarshal carica i dati da un formato .properties in una struct
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
						// Gestisci slice
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
						// Gestisci mappe
						mapKey := extractMapKey(subKey) // Prendi solo la chiave
						if mapKey != "" {
							// Assicurati che la mappa sia inizializzata
							if fieldValue.IsNil() {
								fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
							}
							// Creare o aggiornare la chiave nella mappa
							mapValue := fieldValue.MapIndex(reflect.ValueOf(mapKey))
							if !mapValue.IsValid() {
								// Se non esiste, crea un nuovo valore per la chiave
								mapValue = reflect.New(fieldValue.Type().Elem()).Elem()
							}
							// Imposta il valore nella mappa
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
		// Gestisci mappe
		mapKey := extractMapKey(key)
		if mapKey != "" {
			// Assicurati che la mappa sia inizializzata
			if val.IsNil() {
				val.Set(reflect.MakeMap(val.Type()))
			}
			mapValue := val.MapIndex(reflect.ValueOf(mapKey))
			if !mapValue.IsValid() {
				// Se non esiste, crea un nuovo valore
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

// extractMapKey estrae la chiave della mappa
func extractMapKey(key string) string {
	parts := strings.SplitN(key, ".", 2)
	return parts[0] // Ritorna solo la chiave
}

// setBasicValue imposta il valore di un campo di base
func setBasicValue(v reflect.Value, value string) error {
	if !v.IsValid() {
		return fmt.Errorf("invalid value provided")
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
