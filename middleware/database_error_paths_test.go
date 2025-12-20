package middleware

import (
	"strings"
	"testing"
)

func TestGetCurrentSuiteID_NoCurrentSuite_SetsDefault(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("UPDATE suites SET is_current = 0"); err != nil {
		t.Fatalf("clear current suite: %v", err)
	}

	gotID, err := GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("GetCurrentSuiteID failed: %v", err)
	}
	if gotID <= 0 {
		t.Fatalf("expected positive suite id, got %d", gotID)
	}

	var defaultCurrent bool
	if err := db.QueryRow("SELECT is_current FROM suites WHERE name = 'default'").Scan(&defaultCurrent); err != nil {
		t.Fatalf("query default suite: %v", err)
	}
	if !defaultCurrent {
		t.Fatalf("expected default suite to be set current")
	}
}

func TestGetCurrentSuiteID_UpdateDefaultError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("UPDATE suites SET is_current = 0"); err != nil {
		t.Fatalf("clear current suite: %v", err)
	}

	if _, err := db.Exec(`CREATE TRIGGER abort_suite_update
		BEFORE UPDATE ON suites
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	_, err := GetCurrentSuiteID()
	if err == nil {
		t.Fatalf("expected GetCurrentSuiteID to return an error when default update fails")
	}
	if !strings.Contains(err.Error(), "failed to set default suite as current") {
		t.Fatalf("expected default-update error, got %v", err)
	}
}

func TestGetSuiteID_InsertError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec(`CREATE TRIGGER abort_suite_insert
		BEFORE INSERT ON suites
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	_, err := GetSuiteID("new-suite")
	if err == nil {
		t.Fatalf("expected GetSuiteID to return an error when insert fails")
	}
	if !strings.Contains(err.Error(), "failed to create suite") {
		t.Fatalf("expected create-suite error, got %v", err)
	}
}

func TestListSuites_ScanError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE suites"); err != nil {
		t.Fatalf("drop suites: %v", err)
	}
	if _, err := db.Exec("CREATE VIEW suites AS SELECT NULL AS name"); err != nil {
		t.Fatalf("create suites view: %v", err)
	}

	_, err := ListSuites()
	if err == nil {
		t.Fatalf("expected ListSuites to return an error when scan fails")
	}
	if !strings.Contains(err.Error(), "failed to scan suite name") {
		t.Fatalf("expected scan error, got %v", err)
	}
}

func TestDeleteSuite_DeleteExecError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := GetSuiteID("to-delete"); err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	if _, err := db.Exec(`CREATE TRIGGER abort_suite_delete
		BEFORE DELETE ON suites
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	err := DeleteSuite("to-delete")
	if err == nil {
		t.Fatalf("expected DeleteSuite to return an error when delete fails")
	}
	if !strings.Contains(err.Error(), "failed to delete suite") {
		t.Fatalf("expected delete-suite error, got %v", err)
	}
}

func TestSetCurrentSuite_ClearFlagError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec(`CREATE TRIGGER abort_suite_update
		BEFORE UPDATE ON suites
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	err := SetCurrentSuite("test-suite")
	if err == nil {
		t.Fatalf("expected SetCurrentSuite to return an error when update fails")
	}
	if !strings.Contains(err.Error(), "failed to clear current suite flag") {
		t.Fatalf("expected clear-flag error, got %v", err)
	}
}

func TestDeleteSuite_UpdateDefaultError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := GetSuiteID("current-suite"); err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	// Mark the suite as current before installing the trigger.
	if _, err := db.Exec("UPDATE suites SET is_current = 0"); err != nil {
		t.Fatalf("clear current suite: %v", err)
	}
	if _, err := db.Exec("UPDATE suites SET is_current = 1 WHERE name = 'current-suite'"); err != nil {
		t.Fatalf("set current suite: %v", err)
	}

	if _, err := db.Exec(`CREATE TRIGGER abort_suite_update
		BEFORE UPDATE ON suites
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	err := DeleteSuite("current-suite")
	if err == nil {
		t.Fatalf("expected DeleteSuite to return an error when default update fails")
	}
	if !strings.Contains(err.Error(), "failed to set default suite as current") {
		t.Fatalf("expected default-update error, got %v", err)
	}
}

func TestRenameSuite_ExistenceCheckError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := CloseDB(); err != nil {
		t.Fatalf("CloseDB failed: %v", err)
	}

	err := RenameSuite("old-name", "new-name")
	if err == nil {
		t.Fatalf("expected RenameSuite to return an error when database is closed")
	}
	if !strings.Contains(err.Error(), "failed to check if suite exists") {
		t.Fatalf("expected existence-check error, got %v", err)
	}
}

func TestSetCurrentSuite_SetCurrentError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec(`CREATE TRIGGER abort_set_current_suite
		BEFORE UPDATE ON suites
		WHEN NEW.is_current = 1
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	err := SetCurrentSuite("test-suite")
	if err == nil {
		t.Fatalf("expected SetCurrentSuite to return an error when setting current fails")
	}
	if !strings.Contains(err.Error(), "failed to set current suite") {
		t.Fatalf("expected set-current error, got %v", err)
	}
}

func TestListSuites_QueryError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE suites"); err != nil {
		t.Fatalf("drop suites: %v", err)
	}

	_, err := ListSuites()
	if err == nil {
		t.Fatalf("expected ListSuites to return an error when suites table is missing")
	}
	if !strings.Contains(err.Error(), "failed to query suites") {
		t.Fatalf("expected query error, got %v", err)
	}
}

func TestGetSuiteID_QueryError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE suites"); err != nil {
		t.Fatalf("drop suites: %v", err)
	}

	if _, err := GetSuiteID("any-suite"); err == nil {
		t.Fatalf("expected GetSuiteID to return an error when suites table is missing")
	}
}

func TestSetCurrentSuite_GetSuiteIDError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE suites"); err != nil {
		t.Fatalf("drop suites: %v", err)
	}

	err := SetCurrentSuite("any-suite")
	if err == nil {
		t.Fatalf("expected SetCurrentSuite to return an error when suites table is missing")
	}
	if !strings.Contains(err.Error(), "failed to get suite") {
		t.Fatalf("expected GetSuiteID-wrapped error, got %v", err)
	}
}
