package main

import (
	"os"
	"strings"
	"testing"
)

func cssBlock(t *testing.T, css, selector string) string {
	t.Helper()

	start := strings.Index(css, selector)
	if start < 0 {
		t.Fatalf("templates/arena.css missing selector %q", selector)
	}

	open := strings.Index(css[start:], "{")
	if open < 0 {
		t.Fatalf("templates/arena.css selector %q missing '{'", selector)
	}
	open += start

	close := strings.Index(css[open:], "}")
	if close < 0 {
		t.Fatalf("templates/arena.css selector %q missing closing '}'", selector)
	}
	close += open

	return css[open : close+1]
}

func TestArenaCSS_DefinesCompactSidebarWidthVariable(t *testing.T) {
	t.Helper()

	cssBytes, err := os.ReadFile("templates/arena.css")
	if err != nil {
		t.Fatalf("read templates/arena.css: %v", err)
	}
	css := string(cssBytes)

	if !strings.Contains(css, "--sidebar-width:") {
		t.Fatalf("templates/arena.css must define --sidebar-width for a thinner nav rail")
	}

	shell := cssBlock(t, css, ".arena-shell")
	if !strings.Contains(shell, "grid-template-columns: var(--sidebar-width) 1fr") {
		t.Fatalf("templates/arena.css .arena-shell must use --sidebar-width in grid-template-columns")
	}
}

func TestArenaCSS_StylesDropdownsAndFileInputs(t *testing.T) {
	t.Helper()

	cssBytes, err := os.ReadFile("templates/arena.css")
	if err != nil {
		t.Fatalf("read templates/arena.css: %v", err)
	}
	css := string(cssBytes)

	if !strings.Contains(css, "select") {
		t.Fatalf("templates/arena.css must style <select> dropdowns")
	}
	if !strings.Contains(css, `input[type="file"]`) {
		t.Fatalf("templates/arena.css must style file inputs (input[type=\"file\"])")
	}
}

func TestArenaCSS_ProvidesCompactHeaderAndToolbarLayouts(t *testing.T) {
	t.Helper()

	cssBytes, err := os.ReadFile("templates/arena.css")
	if err != nil {
		t.Fatalf("read templates/arena.css: %v", err)
	}
	css := string(cssBytes)

	for _, selector := range []string{
		".sticky-header",
		".sticky-footer",
		".title-row",
		".filter-form",
		".search-form",
		".flex-form",
	} {
		_ = cssBlock(t, css, selector)
	}

	titleRow := cssBlock(t, css, ".title-row")
	if !strings.Contains(titleRow, "display: flex") {
		t.Fatalf("templates/arena.css .title-row must be a flex row so controls can share a line when space allows")
	}
}

func TestArenaCSS_ScrollButtonsUseSidebarSpace(t *testing.T) {
	t.Helper()

	cssBytes, err := os.ReadFile("templates/arena.css")
	if err != nil {
		t.Fatalf("read templates/arena.css: %v", err)
	}
	css := string(cssBytes)

	block := cssBlock(t, css, ".scroll-buttons")
	if !strings.Contains(block, "left:") {
		t.Fatalf("templates/arena.css .scroll-buttons must be anchored on the left to utilize sidebar space")
	}
	if strings.Contains(block, "right:") {
		t.Fatalf("templates/arena.css .scroll-buttons should not be anchored on the right anymore")
	}
}

