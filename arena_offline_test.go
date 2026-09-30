package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTemplates_NoCDNScripts enforces the fully-offline claim: every script
// and stylesheet a page loads must be served by this server, not a CDN.
func TestTemplates_NoCDNScripts(t *testing.T) {
	entries, err := os.ReadDir("templates")
	if err != nil {
		t.Fatalf("read templates dir: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".html") {
			continue
		}
		b, err := os.ReadFile(filepath.Join("templates", name))
		if err != nil {
			t.Fatalf("read templates/%s: %v", name, err)
		}
		for _, line := range strings.Split(string(b), "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "<script") && !strings.HasPrefix(trimmed, "<link") {
				continue
			}
			if strings.Contains(trimmed, "cdn.") {
				t.Errorf("templates/%s loads an external CDN resource, breaking the offline claim: %s", name, trimmed)
			}
		}
	}
}
