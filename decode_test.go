package jprop

import (
	"reflect"
	"testing"
)

// testStruct is used for unmarshalling tests
type testStruct struct {
	Name   string            `jprop:"name"`
	Age    int               `jprop:"age"`
	Active bool              `jprop:"active"`
	Tags   []string          `jprop:"tags"`
	Props  map[string]string `jprop:"props"`
}

func TestUnmarshal(t *testing.T) {
	type args struct {
		data []byte
		v    interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Basic fields",
			args: args{
				data: []byte(`name=John Doe
age=30
active=true`),
				v: &testStruct{},
			},
			want: &testStruct{
				Name:   "John Doe",
				Age:    30,
				Active: true,
			},
			wantErr: false,
		},
		{
			name: "Slice fields",
			args: args{
				data: []byte(`name=John Doe
tags=go,programming,testing`),
				v: &testStruct{},
			},
			want: &testStruct{
				Name: "John Doe",
				Tags: []string{"go", "programming", "testing"},
			},
			wantErr: false,
		},
		{
			name: "Map fields",
			args: args{
				data: []byte(`name=John Doe
props.language=go
props.editor=vscode`),
				v: &testStruct{},
			},
			want: &testStruct{
				Name: "John Doe",
				Props: map[string]string{
					"language": "go",
					"editor":   "vscode",
				},
			},
			wantErr: false,
		},
		{
			name: "Map with empty key",
			args: args{
				data: []byte(`props.=emptyValue`), // Test for empty map key handling
				v:    &testStruct{Props: make(map[string]string)},
			},
			want: &testStruct{
				Props: map[string]string{
					"": "emptyValue", // Key is empty, should store as such
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid integer",
			args: args{
				data: []byte(`age=invalid`),
				v:    &testStruct{},
			},
			want:    &testStruct{},
			wantErr: true,
		},
		{
			name: "Invalid boolean",
			args: args{
				data: []byte(`active=notabool`),
				v:    &testStruct{},
			},
			want:    &testStruct{},
			wantErr: true,
		},
		{
			name: "Empty fields",
			args: args{
				data: []byte(`name=
tags=`),
				v: &testStruct{},
			},
			want: &testStruct{
				Name: "",
				Tags: []string{""}, // Expected empty tag as per input
			},
			wantErr: false,
		},
		{
			name: "Empty input data",
			args: args{
				data: []byte(``),
				v:    &testStruct{},
			},
			want:    &testStruct{},
			wantErr: false,
		},
		{
			name: "Commented lines",
			args: args{
				data: []byte(`# This is a comment
				name=John Doe
				# Another comment
				age=25
				active=true`),
				v: &testStruct{},
			},
			want: &testStruct{
				Name:   "John Doe",
				Age:    25,
				Active: true,
			},
			wantErr: false,
		},
		{
			name: "Nested struct with missing subKey",
			args: args{
				data: []byte(`props.key=value`),
				v: &testStruct{
					Props: nil,
				},
			},
			want: &testStruct{
				Props: map[string]string{
					"key": "value",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid slice structure",
			args: args{
				data: []byte(`tags=invalid,slice`),
				v:    &testStruct{},
			},
			want: &testStruct{
				Tags: []string{"invalid", "slice"},
			},
			wantErr: false,
		},
		{
			name: "Non-existent field in struct",
			args: args{
				data: []byte(`nonexistent=field`),
				v:    &testStruct{},
			},
			want:    &testStruct{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unmarshal(tt.args.data, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() [%s] error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			// Compare result with expected
			if !tt.wantErr && !reflect.DeepEqual(tt.args.v, tt.want) {
				t.Errorf("Unmarshal() [%s] = %+v, want %+v", tt.name, tt.args.v, tt.want)
			}
		})
	}
}
