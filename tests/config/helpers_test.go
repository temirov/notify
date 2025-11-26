package tests

import (
	"os"
	"path/filepath"
	"testing"
)

func repoPath(pathSegments ...string) string {
	fullSegments := append([]string{"..", ".."}, pathSegments...)
	return filepath.Join(fullSegments...)
}

func readRepoFile(t *testing.T, pathSegments ...string) []byte {
	t.Helper()

	fullPath := repoPath(pathSegments...)

	data, readErr := os.ReadFile(fullPath)
	if readErr != nil {
		t.Fatalf("failed to read %s: %v", fullPath, readErr)
	}

	return data
}
