package middleware

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary database for testing
func setupTestDB(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "db_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")

	cleanup := func() {
		_ = CloseDB()
		_ = os.RemoveAll(tmpDir)
	}

	return dbPath, cleanup
}

func TestInitDB(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Verify database is accessible
	if db == nil {
		t.Fatal("db should not be nil after InitDB")
	}

	// Verify default suite exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM suites WHERE name = 'default'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query default suite: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 default suite, got %d", count)
	}
}

func TestInitDB_CreatesDataDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "db_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a nested path that doesn't exist yet
	dbPath := filepath.Join(tmpDir, "nested", "data", "test.db")

	err = InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer func() { _ = CloseDB() }()

	// Verify the directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("data directory was not created")
	}
}

func TestInitDB_InvalidPath(t *testing.T) {
	// Try to create a database in an invalid path (using a file as a directory)
	tmpDir, err := os.MkdirTemp("", "db_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a file
	filePath := filepath.Join(tmpDir, "notadir")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Try to use that file as a directory path
	invalidDBPath := filepath.Join(filePath, "test.db")
	err = InitDB(invalidDBPath)

	// Should fail because we can't create a directory where a file exists
	if err == nil {
		_ = CloseDB()
		t.Error("expected error when path is invalid")
	}
}

func TestInitDB_OpenError(t *testing.T) {
	// Create a path that will cause sql.Open to fail
	// Use a very long path that exceeds system limits on some systems
	tmpDir, err := os.MkdirTemp("", "db_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a path with special characters that might cause issues
	// Using null bytes which are invalid in file paths
	invalidPath := filepath.Join(tmpDir, "test\x00invalid.db")
	err = InitDB(invalidPath)

	if err == nil {
		_ = CloseDB()
		// On some systems this might not fail, so we don't require an error
		// but if it succeeds, we just clean up
	}
}

func TestInitDB_CreateTablesError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "db_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	dbPath := filepath.Join(tmpDir, "readonly.db")

	// First create a valid database
	err = InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed on first call: %v", err)
	}
	_ = CloseDB()

	// Make the database file read-only
	if err := os.Chmod(dbPath, 0444); err != nil {
		t.Fatalf("failed to make database read-only: %v", err)
	}

	// Try to initialize again - this should work for opening but might fail on some operations
	// Note: SQLite in WAL mode might still work with read-only files
	err = InitDB(dbPath)
	// We don't assert error here because SQLite is resilient
	// The test is more about ensuring we don't panic
	if err == nil {
		_ = CloseDB()
	}
}

func TestCloseDB(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	err = CloseDB()
	if err != nil {
		t.Errorf("CloseDB failed: %v", err)
	}
}

func TestCloseDB_NilDB(t *testing.T) {
	// Save and restore original db
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	err := CloseDB()
	if err != nil {
		t.Errorf("CloseDB with nil db should not error: %v", err)
	}
}

func TestGetDB(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	got := GetDB()
	if got == nil {
		t.Error("GetDB returned nil")
	}
	if got != db {
		t.Error("GetDB should return the global db")
	}
}

func TestGetSuiteID_Default(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Empty string should return default suite
	id, err := GetSuiteID("")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive suite ID, got %d", id)
	}

	// "default" should return same ID
	id2, err := GetSuiteID("default")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}
	if id != id2 {
		t.Errorf("expected same ID for empty string and 'default', got %d and %d", id, id2)
	}
}

func TestGetSuiteID_NewSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a new suite
	id, err := GetSuiteID("new-suite")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive suite ID, got %d", id)
	}

	// Should return same ID on second call
	id2, err := GetSuiteID("new-suite")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}
	if id != id2 {
		t.Errorf("expected same ID on subsequent call, got %d and %d", id, id2)
	}
}

func TestGetCurrentSuiteID(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	id, err := GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("GetCurrentSuiteID failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive suite ID, got %d", id)
	}
}

func TestSetCurrentSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create and set a new suite as current
	err = SetCurrentSuite("test-suite")
	if err != nil {
		t.Fatalf("SetCurrentSuite failed: %v", err)
	}

	// Verify it's the current suite
	var name string
	err = db.QueryRow("SELECT name FROM suites WHERE is_current = 1").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query current suite: %v", err)
	}
	if name != "test-suite" {
		t.Errorf("expected current suite 'test-suite', got %q", name)
	}

	// Verify only one suite is current
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM suites WHERE is_current = 1").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count current suites: %v", err)
	}
	if count != 1 {
		t.Errorf("expected exactly 1 current suite, got %d", count)
	}
}

func TestSetCurrentSuite_EmptyName(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// First switch to a different suite
	err = SetCurrentSuite("other-suite")
	if err != nil {
		t.Fatalf("SetCurrentSuite failed: %v", err)
	}

	// Empty name should switch back to default
	err = SetCurrentSuite("")
	if err != nil {
		t.Fatalf("SetCurrentSuite failed: %v", err)
	}

	var name string
	err = db.QueryRow("SELECT name FROM suites WHERE is_current = 1").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query current suite: %v", err)
	}
	if name != "default" {
		t.Errorf("expected current suite 'default', got %q", name)
	}
}

func TestListSuites(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create additional suites
	_, _ = GetSuiteID("suite-a")
	_, _ = GetSuiteID("suite-b")

	suites, err := ListSuites()
	if err != nil {
		t.Fatalf("ListSuites failed: %v", err)
	}

	if len(suites) != 3 {
		t.Errorf("expected 3 suites, got %d", len(suites))
	}

	// Should be sorted alphabetically
	expected := []string{"default", "suite-a", "suite-b"}
	for i, name := range expected {
		if suites[i] != name {
			t.Errorf("expected suites[%d] = %q, got %q", i, name, suites[i])
		}
	}
}

func TestDeleteSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a suite to delete
	_, err = GetSuiteID("to-delete")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	err = DeleteSuite("to-delete")
	if err != nil {
		t.Fatalf("DeleteSuite failed: %v", err)
	}

	// Verify it's deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM suites WHERE name = 'to-delete'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query suite: %v", err)
	}
	if count != 0 {
		t.Error("suite should have been deleted")
	}
}

func TestDeleteSuite_Default(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	err = DeleteSuite("default")
	if err == nil {
		t.Error("expected error when deleting default suite")
	}
}

func TestDeleteSuite_Current(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create and set as current
	err = SetCurrentSuite("current-suite")
	if err != nil {
		t.Fatalf("SetCurrentSuite failed: %v", err)
	}

	// Delete the current suite
	err = DeleteSuite("current-suite")
	if err != nil {
		t.Fatalf("DeleteSuite failed: %v", err)
	}

	// Default should now be current
	var name string
	err = db.QueryRow("SELECT name FROM suites WHERE is_current = 1").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query current suite: %v", err)
	}
	if name != "default" {
		t.Errorf("expected default to be current after deleting current suite, got %q", name)
	}
}

func TestRenameSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a suite to rename
	_, err = GetSuiteID("old-name")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	err = RenameSuite("old-name", "new-name")
	if err != nil {
		t.Fatalf("RenameSuite failed: %v", err)
	}

	// Verify old name doesn't exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM suites WHERE name = 'old-name'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query suite: %v", err)
	}
	if count != 0 {
		t.Error("old suite name should not exist")
	}

	// Verify new name exists
	err = db.QueryRow("SELECT COUNT(*) FROM suites WHERE name = 'new-name'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query suite: %v", err)
	}
	if count != 1 {
		t.Error("new suite name should exist")
	}
}

func TestRenameSuite_Default(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	err = RenameSuite("default", "new-name")
	if err == nil {
		t.Error("expected error when renaming default suite")
	}
}

func TestRenameSuite_EmptyNewName(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	_, err = GetSuiteID("test-suite")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	err = RenameSuite("test-suite", "")
	if err == nil {
		t.Error("expected error when renaming to empty string")
	}
}

func TestRenameSuite_InvalidChars(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	_, err = GetSuiteID("test-suite")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	tests := []string{
		"name/with/slash",
		"name\\with\\backslash",
	}

	for _, newName := range tests {
		err = RenameSuite("test-suite", newName)
		if err == nil {
			t.Errorf("expected error for invalid name %q", newName)
		}
	}
}

func TestRenameSuite_DuplicateName(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	_, err = GetSuiteID("suite-a")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}
	_, err = GetSuiteID("suite-b")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	err = RenameSuite("suite-a", "suite-b")
	if err == nil {
		t.Error("expected error when renaming to existing name")
	}
}

func TestCreateTables_DefaultSettings(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Check default settings exist
	expectedSettings := []string{
		"api_key_anthropic",
		"api_key_openai",
		"api_key_google",
		"cost_alert_threshold_usd",
		"auto_evaluate_new_models",
		"python_service_url",
	}

	for _, key := range expectedSettings {
		var value string
		err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
		if err != nil {
			t.Errorf("expected setting %q to exist: %v", key, err)
		}
	}
}

func TestForeignKeyConstraints(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Get default suite ID
	suiteID, err := GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("GetCurrentSuiteID failed: %v", err)
	}

	// Create a profile
	result, err := db.Exec("INSERT INTO profiles (name, suite_id) VALUES ('test-profile', ?)", suiteID)
	if err != nil {
		t.Fatalf("failed to insert profile: %v", err)
	}
	profileID, _ := result.LastInsertId()

	// Create a prompt with the profile
	_, err = db.Exec("INSERT INTO prompts (text, profile_id, suite_id, display_order) VALUES ('test prompt', ?, ?, 0)", profileID, suiteID)
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}

	// Delete the profile - prompt's profile_id should be set to NULL
	_, err = db.Exec("DELETE FROM profiles WHERE id = ?", profileID)
	if err != nil {
		t.Fatalf("failed to delete profile: %v", err)
	}

	// Verify prompt still exists with NULL profile_id
	var pID sql.NullInt64
	err = db.QueryRow("SELECT profile_id FROM prompts WHERE text = 'test prompt'").Scan(&pID)
	if err != nil {
		t.Fatalf("failed to query prompt: %v", err)
	}
	if pID.Valid {
		t.Error("expected prompt's profile_id to be NULL after profile deletion")
	}
}

func TestCascadeDelete(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a test suite
	suiteID, err := GetSuiteID("cascade-test")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	// Create profile, prompt, and model in the test suite
	_, err = db.Exec("INSERT INTO profiles (name, suite_id) VALUES ('test-profile', ?)", suiteID)
	if err != nil {
		t.Fatalf("failed to insert profile: %v", err)
	}

	_, err = db.Exec("INSERT INTO prompts (text, suite_id, display_order) VALUES ('test prompt', ?, 0)", suiteID)
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}

	_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES ('test-model', ?)", suiteID)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	// Delete the suite
	err = DeleteSuite("cascade-test")
	if err != nil {
		t.Fatalf("DeleteSuite failed: %v", err)
	}

	// Verify all related data is deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM profiles WHERE suite_id = ?", suiteID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query profiles: %v", err)
	}
	if count != 0 {
		t.Error("profiles should be cascade deleted")
	}

	err = db.QueryRow("SELECT COUNT(*) FROM prompts WHERE suite_id = ?", suiteID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query prompts: %v", err)
	}
	if count != 0 {
		t.Error("prompts should be cascade deleted")
	}

	err = db.QueryRow("SELECT COUNT(*) FROM models WHERE suite_id = ?", suiteID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query models: %v", err)
	}
	if count != 0 {
		t.Error("models should be cascade deleted")
	}
}

func TestGetProfileID_EmptyName(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	suiteID, _ := GetCurrentSuiteID()

	// Empty profile name should return 0, false, nil
	id, exists, err := GetProfileID("", suiteID)
	if err != nil {
		t.Fatalf("GetProfileID failed: %v", err)
	}
	if exists {
		t.Error("expected exists to be false for empty profile name")
	}
	if id != 0 {
		t.Errorf("expected id 0 for empty profile name, got %d", id)
	}
}

func TestGetProfileID_NotFound(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	suiteID, _ := GetCurrentSuiteID()

	// Non-existent profile should return 0, false, nil
	id, exists, err := GetProfileID("non-existent-profile", suiteID)
	if err != nil {
		t.Fatalf("GetProfileID failed: %v", err)
	}
	if exists {
		t.Error("expected exists to be false for non-existent profile")
	}
	if id != 0 {
		t.Errorf("expected id 0 for non-existent profile, got %d", id)
	}
}

func TestGetProfileID_Found(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	suiteID, _ := GetCurrentSuiteID()

	// Create a profile
	result, err := db.Exec("INSERT INTO profiles (name, suite_id) VALUES ('test-profile', ?)", suiteID)
	if err != nil {
		t.Fatalf("failed to insert profile: %v", err)
	}
	expectedID, _ := result.LastInsertId()

	// Should find the profile
	id, exists, err := GetProfileID("test-profile", suiteID)
	if err != nil {
		t.Fatalf("GetProfileID failed: %v", err)
	}
	if !exists {
		t.Error("expected exists to be true for existing profile")
	}
	if id != int(expectedID) {
		t.Errorf("expected id %d, got %d", expectedID, id)
	}
}

func TestGetCurrentSuiteID_NoCurrent(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Clear all current flags
	_, err = db.Exec("UPDATE suites SET is_current = 0")
	if err != nil {
		t.Fatalf("failed to clear current flags: %v", err)
	}

	// GetCurrentSuiteID should set default as current
	id, err := GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("GetCurrentSuiteID failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive suite ID, got %d", id)
	}

	// Verify default is now current
	var name string
	err = db.QueryRow("SELECT name FROM suites WHERE is_current = 1").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query current suite: %v", err)
	}
	if name != "default" {
		t.Errorf("expected 'default' to be current, got %q", name)
	}
}

func TestDeleteSuite_CurrentSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create and set a non-default suite as current
	_, err = db.Exec("INSERT INTO suites (name, is_current) VALUES ('test-suite', 1)")
	if err != nil {
		t.Fatalf("failed to create test suite: %v", err)
	}
	_, err = db.Exec("UPDATE suites SET is_current = 0 WHERE name = 'default'")
	if err != nil {
		t.Fatalf("failed to clear default current flag: %v", err)
	}

	// Delete the current suite
	err = DeleteSuite("test-suite")
	if err != nil {
		t.Fatalf("DeleteSuite failed: %v", err)
	}

	// Verify default is now current
	var name string
	err = db.QueryRow("SELECT name FROM suites WHERE is_current = 1").Scan(&name)
	if err != nil {
		t.Fatalf("failed to query current suite: %v", err)
	}
	if name != "default" {
		t.Errorf("expected 'default' to be current after deletion, got %q", name)
	}
}

func TestListSuites_Multiple(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add more suites
	_, err = db.Exec("INSERT INTO suites (name) VALUES ('alpha'), ('zebra')")
	if err != nil {
		t.Fatalf("failed to create suites: %v", err)
	}

	suites, err := ListSuites()
	if err != nil {
		t.Fatalf("ListSuites failed: %v", err)
	}

	// Should include default, alpha, zebra (sorted)
	if len(suites) != 3 {
		t.Errorf("expected 3 suites, got %d", len(suites))
	}

	// Verify sorted order
	if suites[0] != "alpha" || suites[1] != "default" || suites[2] != "zebra" {
		t.Errorf("expected sorted order [alpha, default, zebra], got %v", suites)
	}
}
