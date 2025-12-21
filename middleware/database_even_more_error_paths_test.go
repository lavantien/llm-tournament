package middleware

import (
	"strings"
	"testing"
)

func TestDeleteSuite_GetSuiteIDError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := CloseDB(); err != nil {
		t.Fatalf("CloseDB failed: %v", err)
	}

	err := DeleteSuite("not-default")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to get suite ID") {
		t.Fatalf("expected GetSuiteID error, got %v", err)
	}
}

func TestRenameSuite_UpdateError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := GetSuiteID("old-name"); err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	if _, err := db.Exec(`CREATE TRIGGER abort_suite_rename
			BEFORE UPDATE ON suites
			BEGIN
				SELECT RAISE(FAIL, 'nope');
			END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	err := RenameSuite("old-name", "new-name")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to rename suite") {
		t.Fatalf("expected rename error, got %v", err)
	}
}
