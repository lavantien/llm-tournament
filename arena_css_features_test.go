package main

import (
	"os"
	"strings"
	"testing"
)

func TestArenaCSS_HidesHiddenData(t *testing.T) {
	t.Helper()

	cssBytes, err := os.ReadFile("templates/arena.css")
	if err != nil {
		t.Fatalf("read templates/arena.css: %v", err)
	}
	css := string(cssBytes)

	if !strings.Contains(css, ".hidden-data") {
		t.Fatalf("templates/arena.css missing .hidden-data rule")
	}
	if !strings.Contains(css, "display: none") {
		t.Fatalf("templates/arena.css expected to hide .hidden-data (display: none)")
	}
}

func TestArenaCSS_DefinesScoreClasses(t *testing.T) {
	t.Helper()

	cssBytes, err := os.ReadFile("templates/arena.css")
	if err != nil {
		t.Fatalf("read templates/arena.css: %v", err)
	}
	css := string(cssBytes)

	for _, cls := range []string{
		".score-0",
		".score-20",
		".score-40",
		".score-60",
		".score-80",
		".score-100",
	} {
		if !strings.Contains(css, cls) {
			t.Fatalf("templates/arena.css missing %s", cls)
		}
	}
}
