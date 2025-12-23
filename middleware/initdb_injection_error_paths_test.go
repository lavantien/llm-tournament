package middleware

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestInitDB_OpenError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	original := sqlOpen
	sqlOpen = func(string, string) (*sql.DB, error) { return nil, errors.New("open failed") }
	t.Cleanup(func() { sqlOpen = original })

	err := InitDB(dbPath)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to open database") {
		t.Fatalf("expected open error, got %v", err)
	}
}

func TestInitDB_PragmasError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	original := execPragmas
	execPragmas = func(*sql.DB) error { return errors.New("pragma failed") }
	t.Cleanup(func() { execPragmas = original })

	err := InitDB(dbPath)
	if err == nil {
		_ = CloseDB()
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to set database pragmas") {
		t.Fatalf("expected pragmas error, got %v", err)
	}
}

func TestInitDB_CreateTablesError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	original := createTablesFunc
	createTablesFunc = func() error { return errors.New("create tables failed") }
	t.Cleanup(func() { createTablesFunc = original })

	err := InitDB(dbPath)
	if err == nil {
		_ = CloseDB()
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to create tables") {
		t.Fatalf("expected create tables error, got %v", err)
	}
}
