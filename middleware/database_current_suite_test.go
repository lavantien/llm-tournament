package middleware

import (
	"strings"
	"testing"
)

// With the suites table empty there is no current suite and no default row,
// so GetCurrentSuiteID must still terminate: the old implementation recursed
// into the identical state after a zero-row UPDATE.
func TestGetCurrentSuiteID_EmptySuitesTableTerminates(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if _, err := db.Exec("DELETE FROM suites"); err != nil {
		t.Fatalf("failed to empty suites table: %v", err)
	}

	suiteID, err := GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("expected GetCurrentSuiteID to terminate without error, got %v", err)
	}
	if suiteID <= 0 {
		t.Fatalf("expected a valid suite ID, got %d", suiteID)
	}
	if !SuiteExists("default") {
		t.Error("expected the default suite to be recreated")
	}
	var isCurrent bool
	if err := db.QueryRow("SELECT is_current FROM suites WHERE name = 'default'").Scan(&isCurrent); err != nil {
		t.Fatalf("failed to query default suite: %v", err)
	}
	if !isCurrent {
		t.Error("expected the recreated default suite to be current")
	}
}

func TestGetCurrentSuiteID_InsertDefaultError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if _, err := db.Exec("DELETE FROM suites"); err != nil {
		t.Fatalf("failed to empty suites table: %v", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER abort_current_suite_insert
		BEFORE INSERT ON suites
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("failed to create trigger: %v", err)
	}

	_, err := GetCurrentSuiteID()
	if err == nil {
		t.Fatal("expected error when creating the default suite fails")
	}
	if !strings.Contains(err.Error(), "failed to create default suite") {
		t.Fatalf("expected create default suite error, got %v", err)
	}
}
