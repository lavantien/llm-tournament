package middleware

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestGetSuiteID_LastInsertIDError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	original := lastInsertID
	lastInsertID = func(sql.Result) (int64, error) { return 0, errors.New("boom") }
	t.Cleanup(func() { lastInsertID = original })

	_, err := GetSuiteID("new-suite")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to get suite ID") {
		t.Fatalf("expected last-insert-id error, got %v", err)
	}
}

func TestWriteResults_ModelLastInsertIDError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	original := lastInsertID
	lastInsertID = func(sql.Result) (int64, error) { return 0, errors.New("boom") }
	t.Cleanup(func() { lastInsertID = original })

	err := WriteResults("default", map[string]Result{
		"new-model": {Scores: []int{80}},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to get model ID") {
		t.Fatalf("expected last-insert-id error, got %v", err)
	}
}

