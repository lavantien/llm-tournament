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

	if !strings.Contains(s, "arena-topbar") {
		t.Fatalf("templates/nav.html must include an element with class arena-topbar")
	}
	if !strings.Contains(s, "arena-sidebar") {
		t.Fatalf("templates/nav.html must include an element with class arena-sidebar")
	}
}
