package middleware

import (
	"strings"
	"testing"
)

func TestDeleteSuite_IsCurrentScanError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := GetSuiteID("to-delete"); err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}
	if _, err := db.Exec("UPDATE suites SET is_current = NULL WHERE name = 'to-delete'"); err != nil {
		t.Fatalf("set is_current NULL: %v", err)
	}

	err := DeleteSuite("to-delete")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to check if suite is current") {
		t.Fatalf("expected is-current scan error, got %v", err)
	}
}
