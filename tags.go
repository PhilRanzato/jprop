package jprop

import (
	"reflect"
	"strings"
)

const (
	omitEmptyKey = "omitempty"
)

type tagOptions struct {
	name      string
	omitEmpty bool
}

func parseTagOptions(tag string) tagOptions {
	opts := tagOptions{}
	if tag == "" {
		return opts
	}
	parts := strings.Split(tag, ",")
	opts.name = parts[0]
	for _, opt := range parts[1:] {
		if opt == omitEmptyKey {
			opts.omitEmpty = true
		}
	}
	return opts
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
