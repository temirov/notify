package tests

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

func TestThirdPartyDirectoryRemoved(t *testing.T) {
	t.Helper()

	directoryPath := repoPath("third_party")
	_, statErr := os.Stat(directoryPath)
	if statErr == nil {
		t.Fatalf("third_party directory should not exist (found at %s)", directoryPath)
	}
	if !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatalf("unexpected error when checking %s: %v", directoryPath, statErr)
	}
}
