package main

import (
	"os"
	"strings"
	"testing"
)

func TestREADME_QuickStartDoesNotUseRemovedMakeTargets(t *testing.T) {
	t.Helper()

	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	if !strings.Contains(readme, "## Quick Start") {
		t.Fatalf("README.md missing expected section header %q", "## Quick Start")
	}

	removed := []string{
		"make migrate",
		"make dedup",
		"--migrate-to-sqlite",
		"--cleanup-duplicates",
	}
	for _, token := range removed {
		if strings.Contains(readme, token) {
			t.Errorf("README.md should not reference removed target/flag %q", token)
		}
	}

	if strings.Contains(readme, "[![Coverage](./coverage-badge.svg)]()") {
		t.Errorf("README.md coverage badge should link somewhere (currently empty link)")
	}
}
