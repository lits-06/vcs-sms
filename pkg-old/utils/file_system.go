package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Kiểm tra xem có go.mod file không
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Đã đến root của filesystem
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found")
}
