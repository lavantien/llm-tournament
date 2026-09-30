package handlers

import (
	"llm-tournament/middleware"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// changeToProjectRootResults changes to the project root directory for tests that need templates
func changeToProjectRootResults(t *testing.T) func() {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}
	return func() {
		_ = os.Chdir(originalDir)
	}
}

// setupResultsTestDB creates a test database for results handler tests
func setupResultsTestDB(t *testing.T) func() {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	err := middleware.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	return func() {
		_ = middleware.CloseDB()
	}
}
