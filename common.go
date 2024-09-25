package jprop

import (
	"bufio"
	"strings"
)

type Marshaler interface {
	MarshalProperties() (string, error)
}

type Unmarshaler interface {
	UnmarshalProperties(string) error
}

func parseProperties(data string) map[string]string {
	lines := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		sepIndex := strings.IndexAny(line, "=:")
		if sepIndex < 0 {
			continue
		}
		key := strings.TrimSpace(line[:sepIndex])
		value := strings.TrimSpace(line[sepIndex+1:])
		lines[key] = value
	}
	return lines
}
