package attachments

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitInput(t *testing.T) {
	t.Parallel()

	path, contentType := splitInput(" /tmp/file.txt :: text/plain ")
	if path != "/tmp/file.txt" {
		t.Fatalf("unexpected path %q", path)
	}
	if contentType != "text/plain" {
		t.Fatalf("unexpected content type %q", contentType)
	}

	path, contentType = splitInput("file.bin")
	if path != "file.bin" || contentType != "" {
		t.Fatalf("unexpected result %q %q", path, contentType)
	}
}

func TestLoadInfersContentType(t *testing.T) {
	t.Parallel()

	tempFile := filepath.Join(t.TempDir(), "payload.txt")
	if err := os.WriteFile(tempFile, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	attachments, err := Load([]string{tempFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected one attachment")
	}
	if attachments[0].ContentType == "" {
		t.Fatalf("expected inferred content type")
	}
}

func TestLoadRequiresPath(t *testing.T) {
	t.Parallel()

	_, err := Load([]string{"   "})
	if err == nil {
		t.Fatalf("expected error for missing path")
	}
}
