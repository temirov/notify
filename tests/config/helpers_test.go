package configtest

import (
	"os"
	"path/filepath"
	"testing"
)

func readRepoFile(t *testing.T, pathSegments ...string) []byte {
	t.Helper()

	fullSegments := append([]string{"..", ".."}, pathSegments...)
	fullPath := filepath.Join(fullSegments...)

	data, readErr := os.ReadFile(fullPath)
	if readErr != nil {
		t.Fatalf("failed to read %s: %v", fullPath, readErr)
	}

	return data
}
