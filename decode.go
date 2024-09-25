package jprop

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

func Unmarshal(data []byte, v interface{}) error {
	val := reflect.ValueOf(v).Elem()
	lines := parseProperties(string(data))
	return unmarshalValue(lines, val, "")
}

func unmarshalValue(lines map[string]string, val reflect.Value, prefix string) error {
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
			if fieldValue.CanSet() {
				if fieldValue.Kind() == reflect.Struct {
					err := unmarshalValue(lines, fieldValue, fullKey+".")
					if err != nil {
						return err
					}
				} else {
					lineValue, exists := lines[fullKey]
					if !exists {
						continue
					}
					err := setValueFromString(fieldValue, lineValue)
					if err != nil {
						return err
					}
				}
			}
		}
	case reflect.Map:
		val.Set(reflect.MakeMap(val.Type()))
		for lineKey, lineValue := range lines {
			if strings.HasPrefix(lineKey, prefix) {
				keyPart := strings.TrimPrefix(lineKey, prefix)
				keyParts := strings.SplitN(keyPart, ".", 2)
				mapKey := reflect.New(val.Type().Key()).Elem()
				err := setValueFromString(mapKey, keyParts[0])
				if err != nil {
					return err
				}
				mapValue := reflect.New(val.Type().Elem()).Elem()
				if len(keyParts) > 1 {
					err = unmarshalValue(lines, mapValue, prefix+keyParts[0]+".")
					if err != nil {
						return err
					}
				} else {
					err = setValueFromString(mapValue, lineValue)
					if err != nil {
						return err
					}
				}
				val.SetMapIndex(mapKey, mapValue)
			}
		}
	case reflect.Slice, reflect.Array:
		var maxIndex int
		indices := make(map[int]string)
		for lineKey, lineValue := range lines {
			if strings.HasPrefix(lineKey, prefix[:len(prefix)-1]) {
				keyPart := strings.TrimPrefix(lineKey, prefix[:len(prefix)-1])
				if strings.HasPrefix(keyPart, "[") {
					idxStr := strings.SplitN(keyPart[1:], "]", 2)[0]
					idx, err := strconv.Atoi(idxStr)
					if err != nil {
						return err
					}
					if idx > maxIndex {
						maxIndex = idx
					}
					indices[idx] = lineValue
				}
			}
		}
		sliceValue := reflect.MakeSlice(val.Type(), maxIndex+1, maxIndex+1)
		for idx, strValue := range indices {
			elemValue := sliceValue.Index(idx)
			err := setValueFromString(elemValue, strValue)
			if err != nil {
				return err
			}
		}
		val.Set(sliceValue)
	default:
		lineValue, exists := lines[prefix[:len(prefix)-1]]
		if !exists {
			return nil
		}
		return setValueFromString(val, lineValue)
	}
	return nil
}

func setValueFromString(v reflect.Value, s string) error {
	if v.CanInterface() {
		if um, ok := v.Addr().Interface().(Unmarshaler); ok {
			return um.UnmarshalProperties(s)
		}
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.Complex64, reflect.Complex128:
		c, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return err
		}
		v.SetComplex(c)
	default:
		return errors.New("unsupported type in setValueFromString")
	}
	return nil
}
