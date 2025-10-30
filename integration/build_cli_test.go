package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBuildCLIFromRepositoryRoot(t *testing.T) {
	workingDirectory, workingDirectoryErr := os.Getwd()
	if workingDirectoryErr != nil {
		t.Fatalf("failed to get working directory: %v", workingDirectoryErr)
	}

	repositoryRoot := filepath.Dir(workingDirectory)
	temporaryBinaryDirectory := t.TempDir()
	temporaryBinaryPath := filepath.Join(temporaryBinaryDirectory, "pinguin-cli")

	buildCommand := exec.Command("go", "build", "-o", temporaryBinaryPath, "clients/cli/main.go")
	buildCommand.Dir = repositoryRoot

	commandOutput, buildErr := buildCommand.CombinedOutput()
	if buildErr != nil {
		t.Fatalf("go build failed: %v\n%s", buildErr, string(commandOutput))
	}

	_, binaryStatErr := os.Stat(temporaryBinaryPath)
	if binaryStatErr != nil {
		t.Fatalf("expected binary at %s: %v", temporaryBinaryPath, binaryStatErr)
	}
}
