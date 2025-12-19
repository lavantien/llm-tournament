package main

import (
	"os"
	"strings"
	"testing"
)

func TestAutomatedEvaluationSetupDoc_UsesCurrentCommands(t *testing.T) {
	t.Helper()

	docBytes, err := os.ReadFile("AUTOMATED_EVALUATION_SETUP.md")
	if err != nil {
		t.Fatalf("read AUTOMATED_EVALUATION_SETUP.md: %v", err)
	}
	doc := string(docBytes)

	// Expected to remain stable.
	required := []string{
		"# Automated",
		"python_service",
		"ENCRYPTION_KEY",
		"http://localhost:8001/health",
		"CGO_ENABLED=1 go run .",
	}
	for _, token := range required {
		if !strings.Contains(doc, token) {
			t.Errorf("AUTOMATED_EVALUATION_SETUP.md missing expected content %q", token)
		}
	}

	// These were removed or are no longer recommended.
	disallowed := []string{
		"CGO_ENABLED=1 go run main.go",
		"\nset ENCRYPTION_KEY=",
		"--migrate-to-sqlite",
		"--cleanup-duplicates",
	}
	for _, token := range disallowed {
		if strings.Contains(doc, token) {
			t.Errorf("AUTOMATED_EVALUATION_SETUP.md should not reference %q", token)
		}
	}
}

