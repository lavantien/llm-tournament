package middleware

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestSetCurrentSuite_BeginError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	original := dbBegin
	dbBegin = func() (*sql.Tx, error) { return nil, errors.New("begin failed") }
	t.Cleanup(func() { dbBegin = original })

	err := SetCurrentSuite("default")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to begin transaction") {
		t.Fatalf("expected begin error, got %v", err)
	}
}

func TestDeleteSuite_BeginError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if _, err := GetSuiteID("to-delete"); err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	original := dbBegin
	dbBegin = func() (*sql.Tx, error) { return nil, errors.New("begin failed") }
	t.Cleanup(func() { dbBegin = original })

	err := DeleteSuite("to-delete")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to begin transaction") {
		t.Fatalf("expected begin error, got %v", err)
	}
}

func TestWriteProfileSuite_BeginError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	original := dbBegin
	dbBegin = func() (*sql.Tx, error) { return nil, errors.New("begin failed") }
	t.Cleanup(func() { dbBegin = original })

	err := WriteProfileSuite("default", []Profile{{Name: "p1"}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to begin transaction") {
		t.Fatalf("expected begin error, got %v", err)
	}
}

func TestWritePromptSuite_BeginError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	original := dbBegin
	dbBegin = func() (*sql.Tx, error) { return nil, errors.New("begin failed") }
	t.Cleanup(func() { dbBegin = original })

	err := WritePromptSuite("default", []Prompt{{Text: "p1"}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to begin transaction") {
		t.Fatalf("expected begin error, got %v", err)
	}
}

func TestWriteResults_BeginError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	original := dbBegin
	dbBegin = func() (*sql.Tx, error) { return nil, errors.New("begin failed") }
	t.Cleanup(func() { dbBegin = original })

	err := WriteResults("default", map[string]Result{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to begin transaction") {
		t.Fatalf("expected begin error, got %v", err)
	}
}

func TestReadPromptSuite_RowsErr_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	original := rowsErr
	rowsErr = func(*sql.Rows) error { return errors.New("rows err") }
	t.Cleanup(func() { rowsErr = original })

	_, err := ReadPromptSuite("default")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "error iterating prompt rows") {
		t.Fatalf("expected rows err, got %v", err)
	}
}

func TestWriteResults_PromptRowsErr_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	original := rowsErr
	rowsErr = func(*sql.Rows) error { return errors.New("rows err") }
	t.Cleanup(func() { rowsErr = original })

	err := WriteResults("default", map[string]Result{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "error iterating prompt rows") {
		t.Fatalf("expected rows err, got %v", err)
	}
}

func TestUpdatePromptsOrder_BeginError_DoesNotChangeOrder(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}, {Text: "p2"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	before, err := ReadPromptSuite("default")
	if err != nil {
		t.Fatalf("ReadPromptSuite failed: %v", err)
	}

	original := dbBegin
	dbBegin = func() (*sql.Tx, error) { return nil, errors.New("begin failed") }
	t.Cleanup(func() { dbBegin = original })

	UpdatePromptsOrder([]int{1, 0})

	after, err := ReadPromptSuite("default")
	if err != nil {
		t.Fatalf("ReadPromptSuite failed: %v", err)
	}

	if len(after) != len(before) || after[0].Text != before[0].Text || after[1].Text != before[1].Text {
		t.Fatalf("expected order to remain unchanged, before=%#v after=%#v", before, after)
	}
}

func TestUpdatePromptsOrder_CommitError_DoesNotChangeOrder(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}, {Text: "p2"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	before, err := ReadPromptSuite("default")
	if err != nil {
		t.Fatalf("ReadPromptSuite failed: %v", err)
	}

	original := txCommit
	txCommit = func(*sql.Tx) error { return errors.New("commit failed") }
	t.Cleanup(func() { txCommit = original })

	UpdatePromptsOrder([]int{1, 0})

	after, err := ReadPromptSuite("default")
	if err != nil {
		t.Fatalf("ReadPromptSuite failed: %v", err)
	}

	if len(after) != len(before) || after[0].Text != before[0].Text || after[1].Text != before[1].Text {
		t.Fatalf("expected order to remain unchanged, before=%#v after=%#v", before, after)
	}
}
