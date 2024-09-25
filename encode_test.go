package jprop

import (
	"testing"
)

type testEncoderStruct struct {
	Name    string            `jprop:"name,omitempty"`
	Age     int               `jprop:"age,omitempty"`
	Active  bool              `jprop:"active,omitempty"`
	Tags    []string          `jprop:"tags,omitempty"`
	Props   map[string]string `jprop:"props,omitempty"`
	Address addressStruct     `jprop:"address,omitempty"`
}

type addressStruct struct {
	City    string `jprop:"city,omitempty"`
	Country string `jprop:"country,omitempty"`
}

// TestMarshal tests the Marshal function for various scenarios
func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    testEncoderStruct
		expected string
	}{
		{
			name: "Basic fields",
			input: testEncoderStruct{
				Name:   "John Doe",
				Age:    30,
				Active: true,
			},
			expected: `name=John Doe
age=30
active=true
`,
		},
		{
			name: "Slice fields",
			input: testEncoderStruct{
				Name: "John Doe",
				Tags: []string{"go", "programming", "testing"},
			},
			expected: `name=John Doe
tags=go,programming,testing
`,
		},
		{
			name: "Map fields",
			input: testEncoderStruct{
				Name: "John Doe",
				Props: map[string]string{
					"language": "go",
					"editor":   "vscode",
				},
			},
			expected: `name=John Doe
props.language=go
props.editor=vscode
`,
		},
		{
			name: "Omit empty fields",
			input: testEncoderStruct{
				Name:  "John Doe",
				Props: map[string]string{},
			},
			expected: `name=John Doe
`,
		},
		{
			name: "Nested struct",
			input: testEncoderStruct{
				Name: "John Doe",
				Address: addressStruct{
					City:    "New York",
					Country: "",
				},
			},
			expected: `name=John Doe
address.city=New York
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}
			if string(got) != tt.expected {
				t.Errorf("Marshal() = %v, want %v", string(got), tt.expected)
			}
		})
	}
}
