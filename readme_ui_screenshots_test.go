package main

import (
	"os"
	"strings"
	"testing"
)

func TestREADME_UITourIncludesScreenshots(t *testing.T) {
	t.Helper()

	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	if !strings.Contains(readme, "## UI Tour") {
		t.Fatalf("README.md missing expected section header %q", "## UI Tour")
	}

	required := []string{
		"assets/ui-results.png",
		"assets/ui-prompts.png",
		"assets/ui-edit-prompt.png",
		"assets/ui-profiles.png",
		"assets/ui-evaluate.png",
		"assets/ui-stats.png",
		"assets/ui-settings.png",
	}

	for _, path := range required {
		if !strings.Contains(readme, path) {
			t.Errorf("README.md missing screenshot reference %q", path)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected screenshot file to exist at %q: %v", path, err)
		}
	}
}
