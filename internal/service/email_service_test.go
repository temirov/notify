package service

import (
	"strings"
	"testing"

	"github.com/temirov/pinguin/internal/model"
)

func TestBuildEmailMessageWithoutAttachments(t *testing.T) {
	t.Helper()

	message := buildEmailMessage("from@example.com", "to@example.com", "Subj", "Body", nil)
	if !strings.Contains(message, "Content-Type: text/plain") {
		t.Fatalf("expected plain text content type")
	}
	if strings.Contains(message, "multipart/mixed") {
		t.Fatalf("did not expect multipart headers")
	}
	if !strings.Contains(message, "Body") {
		t.Fatalf("expected body content")
	}
}

func TestBuildEmailMessageWithAttachments(t *testing.T) {
	t.Helper()

	attachment := model.EmailAttachment{
		Filename:    "data.txt",
		ContentType: "text/plain",
		Data:        []byte("hello world"),
	}
	message := buildEmailMessage("from@example.com", "to@example.com", "Subject", "Body", []model.EmailAttachment{attachment})

	if !strings.Contains(message, "multipart/mixed") {
		t.Fatalf("expected multipart content type")
	}
	if !strings.Contains(message, "Content-Disposition: attachment; filename=\"data.txt\"") {
		t.Fatalf("expected content disposition header")
	}
	if !strings.Contains(message, "Content-Transfer-Encoding: base64") {
		t.Fatalf("expected base64 encoding header")
	}
	if !strings.Contains(message, "SGVsbG8gd29ybGQ=") {
		t.Fatalf("expected base64 content in body")
	}
	if !strings.Contains(message, "--PinguinBoundary") {
		t.Fatalf("expected MIME boundary markers")
	}
	if !strings.HasSuffix(strings.TrimSpace(message), "--") {
		t.Fatalf("expected closing boundary terminator")
	}
}
