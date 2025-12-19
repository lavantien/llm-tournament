package main

import (
	"os"
	"strings"
	"testing"
)

func TestMakefile_DoesNotExposeRemovedLegacyTargets(t *testing.T) {
	t.Helper()

	makefileBytes, err := os.ReadFile("makefile")
	if err != nil {
		t.Fatalf("read makefile: %v", err)
	}
	makefile := string(makefileBytes)

	removedTargets := []string{
		"\nmigrate:",
		"\ndedup:",
		"migrate dedup",
		"--migrate-to-sqlite",
		"--cleanup-duplicates",
	}

	for _, token := range removedTargets {
		if strings.Contains(makefile, token) {
			t.Errorf("Makefile should not reference removed legacy target/flag %q", token)
		}
	}

	// Keep a basic sanity check that the repo still advertises the main entrypoints.
	requiredTargets := []string{
		"\nrun:",
		"\ntest:",
		"\nbuild:",
	}
	for _, token := range requiredTargets {
		if !strings.Contains(makefile, token) {
			t.Errorf("Makefile missing expected target %q", strings.TrimSpace(token))
		}
	}
}
