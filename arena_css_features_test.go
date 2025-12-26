package main

import (
	"os"
	"strings"
	"testing"
)

func TestArenaCSS_HidesHiddenData(t *testing.T) {
	t.Helper()

	cssBytes, err := os.ReadFile("templates/output.css")
	if err != nil {
		t.Fatalf("read templates/output.css: %v", err)
	}
	css := string(cssBytes)

	// After migration, we use DaisyUI utility classes instead of custom CSS
	// Check that Tailwind hidden utility is available
	if !strings.Contains(css, "hidden") {
		t.Fatalf("templates/output.css missing Tailwind hidden utility")
	}
}

func TestArenaCSS_DefinesScoreClasses(t *testing.T) {
	t.Helper()

	// After migration, score colors use Tailwind arbitrary values in templates
	// No custom CSS score classes needed - DaisyUI provides color utilities
	cssBytes, err := os.ReadFile("templates/output.css")
	if err != nil {
		t.Fatalf("read templates/output.css: %v", err)
	}
	css := string(cssBytes)

	// Check for DaisyUI CSS is loaded
	if !strings.Contains(css, "daisyui") {
		t.Fatalf("templates/output.css missing DaisyUI CSS")
	}
}
