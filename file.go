package jprop

import (
	"os"
)

func SaveToFile(filename string, v interface{}) error {
	data, err := Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, os.ModePerm)
}

func LoadFromFile(filename string, v interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return Unmarshal(data, v)
}
