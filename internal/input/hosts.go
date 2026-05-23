package input

import (
	"fmt"
	"os"
	"strings"
)

// LoadHosts loads hosts from a file or treats input as a single host.
func LoadHosts(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("hosts: empty input")
	}

	info, err := os.Stat(input)
	if err == nil {
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("hosts: not a regular file: %s", input)
		}
		lines, err := ReadLines(input)
		if err != nil {
			return nil, fmt.Errorf("hosts: %w", err)
		}
		if len(lines) == 0 {
			return nil, fmt.Errorf("hosts: no entries in file: %s", input)
		}
		return lines, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("hosts: %w", err)
	}

	return []string{input}, nil
}
