package tests

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type composeDocument struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	Profiles []string          `yaml:"profiles"`
	Build    *composeBuildSpec `yaml:"build"`
	Image    string            `yaml:"image"`
}

type composeBuildSpec struct {
	Context string `yaml:"context"`
}

func TestComposeProfilesProvideLocalAndImageVariants(t *testing.T) {
	t.Helper()

	documentData := readRepoFile(t, "docker-compose.yaml")

	var document composeDocument
	if unmarshalErr := yaml.Unmarshal(documentData, &document); unmarshalErr != nil {
		t.Fatalf("failed to parse docker-compose.yaml: %v", unmarshalErr)
	}

	localService, localExists := document.Services["pinguin-dev"]
	if !localExists {
		t.Fatalf("compose file missing pinguin-dev service")
	}

	assertProfileContains(t, localService.Profiles, "dev", "pinguin-dev")
	if localService.Build == nil || localService.Build.Context == "" {
		t.Fatalf("pinguin-dev should define a build context for local development")
	}
	if localService.Image != "" {
		t.Fatalf("pinguin-dev should not specify an image because it builds locally")
	}

	imageService, imageExists := document.Services["pinguin"]
	if !imageExists {
		t.Fatalf("compose file missing pinguin service for docker profile")
	}

	assertProfileContains(t, imageService.Profiles, "docker", "pinguin")
	if imageService.Image == "" || !strings.HasPrefix(imageService.Image, "ghcr.io/") {
		t.Fatalf("pinguin docker profile should pull image from ghcr.io, got %q", imageService.Image)
	}
	if imageService.Build != nil {
		t.Fatalf("pinguin docker profile should not include build configuration")
	}
}

func assertProfileContains(t *testing.T, profiles []string, expectedProfile string, serviceName string) {
	t.Helper()

	for _, profile := range profiles {
		if profile == expectedProfile {
			return
		}
	}

	t.Fatalf("%s service is missing %q profile tag", serviceName, expectedProfile)
}
