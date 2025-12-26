package main

import (
	"os"
	"strings"
	"testing"
)

func TestArenaNav_RendersTopbarAndSidebar(t *testing.T) {
	t.Helper()

	b, err := os.ReadFile("templates/nav.html")
	if err != nil {
		t.Fatalf("read templates/nav.html: %v", err)
	}
	s := string(b)

	// Check for DaisyUI components instead of custom classes
	if !strings.Contains(s, "card bg-base-100 shadow-lg") {
		t.Fatalf("templates/nav.html must include DaisyUI card component for topbar")
	}
	if !strings.Contains(s, "card bg-base-100 shadow-lg") {
		t.Fatalf("templates/nav.html must include DaisyUI card component for sidebar")
	}
	// Check for menu component (DaisyUI)
	if !strings.Contains(s, `class="menu`) {
		t.Fatalf("templates/nav.html must include DaisyUI menu component")
	}
}
