package middleware

import (
	"testing"
)

func TestGetCurrentSuiteName(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	name := GetCurrentSuiteName()
	if name != "default" {
		t.Errorf("expected 'default', got %q", name)
	}
}

func TestGetCurrentSuiteName_AfterSwitch(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Switch to a new suite
	err = SetCurrentSuite("test-suite")
	if err != nil {
		t.Fatalf("SetCurrentSuite failed: %v", err)
	}

	name := GetCurrentSuiteName()
	if name != "test-suite" {
		t.Errorf("expected 'test-suite', got %q", name)
	}
}

func TestGetCurrentSuiteName_NoCurrentSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Clear all is_current flags
	_, err = db.Exec("UPDATE suites SET is_current = 0")
	if err != nil {
		t.Fatalf("failed to clear current suite: %v", err)
	}

	// Function should fall back to default
	name := GetCurrentSuiteName()
	if name != "default" {
		t.Errorf("expected 'default', got %q", name)
	}
}

func TestGetCurrentSuiteName_QueryError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Drop suites table to trigger query error
	_, err = db.Exec("DROP TABLE suites")
	if err != nil {
		t.Fatalf("failed to drop suites table: %v", err)
	}

	// Function should return empty string on query error
	name := GetCurrentSuiteName()
	if name != "" {
		t.Errorf("expected empty string on query error, got %q", name)
	}
}

func TestSuiteExists(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if !SuiteExists("default") {
		t.Error("expected default suite to exist")
	}

	if SuiteExists("nonexistent") {
		t.Error("expected nonexistent suite to not exist")
	}
}

func TestListPromptSuites(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create additional suites
	_, _ = GetSuiteID("suite-a")
	_, _ = GetSuiteID("suite-b")

	suites, err := ListPromptSuites()
	if err != nil {
		t.Fatalf("ListPromptSuites failed: %v", err)
	}

	if len(suites) != 3 {
		t.Errorf("expected 3 suites, got %d", len(suites))
	}
}

func TestDeletePromptSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a new suite
	_, _ = GetSuiteID("test-delete-suite")

	// Verify it exists
	if !SuiteExists("test-delete-suite") {
		t.Fatal("test-delete-suite should exist before deletion")
	}

	// Delete it
	err = DeletePromptSuite("test-delete-suite")
	if err != nil {
		t.Fatalf("DeletePromptSuite failed: %v", err)
	}

	// Verify it's gone
	if SuiteExists("test-delete-suite") {
		t.Error("test-delete-suite should not exist after deletion")
	}
}

func TestDeleteProfileSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a suite and add profiles
	_, _ = GetSuiteID("profile-delete-suite")
	err = WriteProfileSuite("profile-delete-suite", []Profile{
		{Name: "Profile to Delete"},
	})
	if err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}

	// Delete it
	err = DeleteProfileSuite("profile-delete-suite")
	if err != nil {
		t.Fatalf("DeleteProfileSuite failed: %v", err)
	}
}

func TestListProfileSuites(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create additional suites
	_, _ = GetSuiteID("profile-suite-a")
	_, _ = GetSuiteID("profile-suite-b")

	suites, err := ListProfileSuites()
	if err != nil {
		t.Fatalf("ListProfileSuites failed: %v", err)
	}

	if len(suites) != 3 {
		t.Errorf("expected 3 suites, got %d", len(suites))
	}
}

func TestRenameSuiteFiles(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a suite
	_, _ = GetSuiteID("old-suite-name")
	if !SuiteExists("old-suite-name") {
		t.Fatal("old-suite-name should exist before rename")
	}

	// Rename it
	err = RenameSuiteFiles("old-suite-name", "new-suite-name")
	if err != nil {
		t.Fatalf("RenameSuiteFiles failed: %v", err)
	}

	// Verify old name is gone and new name exists
	if SuiteExists("old-suite-name") {
		t.Error("old-suite-name should not exist after rename")
	}
	if !SuiteExists("new-suite-name") {
		t.Error("new-suite-name should exist after rename")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 0, -1},
		{0, -1, -1},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.expected {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestGetCurrentSuiteName_DefaultFallback(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Clear is_current flag from all suites
	_, err = db.Exec("UPDATE suites SET is_current = 0")
	if err != nil {
		t.Fatalf("Failed to clear is_current: %v", err)
	}

	// GetCurrentSuiteName should set default as current and return it
	name := GetCurrentSuiteName()
	if name != "default" {
		t.Errorf("expected 'default', got %q", name)
	}

	// Verify is_current was set (is_current is a boolean stored as bool in SQLite)
	var isCurrent bool
	err = db.QueryRow("SELECT is_current FROM suites WHERE name = 'default'").Scan(&isCurrent)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if !isCurrent {
		t.Error("expected is_current to be set to true")
	}
}
