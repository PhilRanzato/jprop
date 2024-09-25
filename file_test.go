package jprop

import (
	"os"
	"reflect"
	"testing"
)

type testFileStruct struct {
	Name   string            `jprop:"name"`
	Age    int               `jprop:"age"`
	Active bool              `jprop:"active"`
	Tags   []string          `jprop:"tags"`
	Props  map[string]string `jprop:"props"`
}

func TestSaveToFile(t *testing.T) {
	testData := &testFileStruct{
		Name:   "John Doe",
		Age:    30,
		Active: true,
		Tags:   []string{"golang", "programming"},
		Props:  map[string]string{"language": "go"},
	}

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "testfile.properties")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the file afterwards

	// Save data to the file
	err = SaveToFile(tempFile.Name(), testData)
	if err != nil {
		t.Errorf("SaveToFile() error = %v", err)
	}

	// Verify the file contents
	data, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	expected := "name=John Doe\nage=30\nactive=true\ntags=golang,programming\nprops.language=go\n"
	if string(data) != expected {
		t.Errorf("SaveToFile() = %s, want %s", string(data), expected)
	}
}

func TestLoadFromFile(t *testing.T) {
	testData := &testFileStruct{
		Name:   "John Doe",
		Age:    30,
		Active: true,
		Tags:   []string{"golang", "programming"},
		Props:  map[string]string{"language": "go"},
	}

	// Create a temporary file and write the properties to it
	tempFile, err := os.CreateTemp("", "testfile.properties")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the file afterwards

	// Write expected data to the file
	expectedData := "name=John Doe\nage=30\nactive=true\ntags=golang,programming\nprops.language=go\n"
	if _, err := tempFile.WriteString(expectedData); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tempFile.Close() // Close the file to flush changes

	// Load data from the file into a new struct
	var loadedData testFileStruct
	err = LoadFromFile(tempFile.Name(), &loadedData)
	if err != nil {
		t.Errorf("LoadFromFile() error = %v", err)
	}

	// Compare loaded data with expected data
	if !reflect.DeepEqual(loadedData, *testData) {
		t.Errorf("LoadFromFile() loaded data = %+v, want %+v", loadedData, *testData)
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	var loadedData testFileStruct
	err := LoadFromFile("non_existent_file.properties", &loadedData)
	if err == nil {
		t.Errorf("LoadFromFile() expected error for non-existent file, got none")
	}
}
