package main

import (
	"os"
	"regexp"
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

func TestArenaNav_NoDuplicateClassAttributes(t *testing.T) {
	b, err := os.ReadFile("templates/nav.html")
	if err != nil {
		t.Fatalf("read templates/nav.html: %v", err)
	}

	// Browsers ignore the second class attribute on an element, so a duplicate
	// silently drops styling. Every anchor must carry at most one class attribute.
	aTag := regexp.MustCompile(`<a\b[^>]*>`)
	for _, tag := range aTag.FindAllString(string(b), -1) {
		if strings.Count(tag, "class=") > 1 {
			t.Errorf("nav anchor has duplicate class attributes, second one is dropped by browsers: %s", tag)
		}
	}
}
