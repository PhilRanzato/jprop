package jprop

import (
	"reflect"
	"testing"
)

func TestParseTagOptions(t *testing.T) {
	tests := []struct {
		tag      string
		expected tagOptions
	}{
		{
			tag:      "name",
			expected: tagOptions{name: "name", omitEmpty: false},
		},
		{
			tag:      "name,omitempty",
			expected: tagOptions{name: "name", omitEmpty: true},
		},
		{
			tag:      "name,omitEmpty",
			expected: tagOptions{name: "name", omitEmpty: false},
		},
		{
			tag:      ",omitempty",
			expected: tagOptions{name: "", omitEmpty: true},
		},
		{
			tag:      "name,extraOption",
			expected: tagOptions{name: "name", omitEmpty: false}, // Extra options are ignored
		},
		{
			tag:      "",
			expected: tagOptions{name: "", omitEmpty: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := parseTagOptions(tt.tag)
			if got != tt.expected {
				t.Errorf("parseTagOptions(%q) = %+v, want %+v", tt.tag, got, tt.expected)
			}
		})
	}
}

func TestIsEmptyValue(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected bool
	}{
		{value: "", expected: true},
		{value: "non-empty", expected: false},
		{value: []string{}, expected: true},
		{value: []string{"item"}, expected: false},
		{value: map[string]string{}, expected: true},
		{value: map[string]string{"key": "value"}, expected: false},
		{value: false, expected: true},
		{value: true, expected: false},
		{value: 0, expected: true},
		{value: 1, expected: false},
		{value: 0.0, expected: true},
		{value: 1.1, expected: false},
		{value: (*string)(nil), expected: true},
		{value: new(string), expected: false},
	}

	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.value).String(), func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			got := isEmptyValue(v)
			if got != tt.expected {
				t.Errorf("isEmptyValue(%v) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}
