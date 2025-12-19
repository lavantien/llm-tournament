package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArenaLayout_AllTemplatesUseShellAndMain(t *testing.T) {
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

		if !strings.Contains(s, `class="arena-shell`) {
			t.Errorf("%s must include arena-shell wrapper", filepath.Join("templates", e.Name()))
		}
		if !strings.Contains(s, `class="arena-main`) {
			t.Errorf("%s must include arena-main wrapper", filepath.Join("templates", e.Name()))
		}
	}
}
