package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArenaTheme_AllTemplatesUseArenaCSS(t *testing.T) {
	t.Helper()

	entries, err := os.ReadDir("templates")
	if err != nil {
		t.Fatalf("read templates dir: %v", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".html" {
			continue
		}
		if e.Name() == "design_preview.html" {
			continue
		}
		if e.Name() == "nav.html" {
			continue
		}

		b, err := os.ReadFile(filepath.Join("templates", e.Name()))
		if err != nil {
			t.Fatalf("read template %s: %v", e.Name(), err)
		}
		s := string(b)

		if !strings.Contains(s, `href="/templates/output.css"`) {
			t.Errorf("%s must reference /templates/output.css", filepath.Join("templates", e.Name()))
		}
	}
}

func TestArenaTheme_AllTemplatesDeclareLanguageAndViewport(t *testing.T) {
	t.Helper()

	entries, err := os.ReadDir("templates")
	if err != nil {
		t.Fatalf("read templates dir: %v", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".html" {
			continue
		}
		if e.Name() == "design_preview.html" || e.Name() == "nav.html" {
			continue
		}

		b, err := os.ReadFile(filepath.Join("templates", e.Name()))
		if err != nil {
			t.Fatalf("read template %s: %v", e.Name(), err)
		}
		s := string(b)

		if !strings.Contains(s, `<html lang="en"`) {
			t.Errorf("%s must declare <html lang=\"en\"> for screen readers", filepath.Join("templates", e.Name()))
		}
		if !strings.Contains(s, `<meta name="viewport" content="width=device-width, initial-scale=1"`) {
			t.Errorf("%s must declare the responsive viewport meta for mobile", filepath.Join("templates", e.Name()))
		}
	}
}
