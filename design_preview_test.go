package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDesignConceptAndPreview_ExistAndStructured(t *testing.T) {
	t.Helper()

	conceptBytes, err := os.ReadFile("DESIGN_CONCEPT.md")
	if err != nil {
		t.Fatalf("DESIGN_CONCEPT.md must exist: %v", err)
	}
	concept := string(conceptBytes)

	for _, needle := range []string{
		"# LLM Tournament Arena — Design Concept",
		"## Color Palette",
		"## Typography",
		"## UI Elements",
		"## Libraries (CDN Only)",
	} {
		if !strings.Contains(concept, needle) {
			t.Errorf("DESIGN_CONCEPT.md missing required section: %q", needle)
		}
	}

	previewPath := filepath.Join("templates", "design_preview.html")
	previewBytes, err := os.ReadFile(previewPath)
	if err != nil {
		t.Fatalf("%s must exist: %v", previewPath, err)
	}
	preview := string(previewBytes)

	for _, needle := range []string{
		"<title>LLM Tournament Arena — Design Preview</title>",
		`href="/assets/favicon.ico"`,
		`src="/assets/logo.webp"`,
		"arena-shell",
		"glass-panel",
		"neon-button",
		"Chart.js",
	} {
		if !strings.Contains(preview, needle) {
			t.Errorf("%s missing required marker: %q", previewPath, needle)
		}
	}

	if strings.Contains(preview, "{{") || strings.Contains(preview, "}}") {
		t.Errorf("%s must be hardcoded HTML (no Go template actions)", previewPath)
	}
}
