package input

import (
	"fmt"
	"os"
)

// LoadPaths loads paths from a file. The input must be an existing regular file.
func LoadPaths(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("paths: not a file: %s", path)
		}
		return nil, fmt.Errorf("paths: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("paths: not a file: %s", path)
	}

	lines, err := ReadLines(path)
	if err != nil {
		return nil, fmt.Errorf("paths: %w", err)
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("paths: no entries in file: %s", path)
	}
	return lines, nil
}
