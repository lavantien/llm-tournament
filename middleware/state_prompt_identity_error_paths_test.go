package middleware

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestWritePromptSuite_LoadIDScanError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	suiteID, err := GetSuiteID("default")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	if _, err := db.Exec("DROP TABLE prompts"); err != nil {
		t.Fatalf("drop prompts: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE prompts (
		id TEXT PRIMARY KEY,
		text TEXT NOT NULL,
		solution TEXT DEFAULT '',
		profile_id INTEGER,
		suite_id INTEGER NOT NULL,
		display_order INTEGER NOT NULL
	)`); err != nil {
		t.Fatalf("create prompts: %v", err)
	}
	if _, err := db.Exec("INSERT INTO prompts (id, text, suite_id, display_order) VALUES (?, ?, ?, ?)", "bad", "p1", suiteID, 0); err != nil {
		t.Fatalf("insert prompt: %v", err)
	}

	err = WritePromptSuite("default", []Prompt{{Text: "p1"}})
	if err == nil {
		t.Fatalf("expected error for unscannable prompt id")
	}
	if !strings.Contains(err.Error(), "failed to load prompt IDs") {
		t.Fatalf("expected load prompt IDs error, got %v", err)
	}
}

func TestWritePromptSuite_LoadIDRowsErr_ReturnsError(t *testing.T) {
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

	prompts := ReadPrompts()
	if err := WritePrompts(prompts); err == nil {
		t.Fatalf("expected error from injected rows iteration failure")
	}
}

func TestWritePromptSuite_UpdateError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}, {Text: "p2"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER abort_prompt_update
		BEFORE UPDATE ON prompts
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	prompts := ReadPrompts()
	err := WritePrompts(prompts)
	if err == nil {
		t.Fatalf("expected error when prompt update fails")
	}
	if !strings.Contains(err.Error(), "failed to update prompt") {
		t.Fatalf("expected update prompt error, got %v", err)
	}
}

func TestWritePromptSuite_DeleteRemovedError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}, {Text: "p2"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER abort_prompt_delete
		BEFORE DELETE ON prompts
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	prompts := ReadPrompts()
	err := WritePrompts(prompts[:1])
	if err == nil {
		t.Fatalf("expected error when removed prompt deletion fails")
	}
	if !strings.Contains(err.Error(), "failed to delete removed prompt") {
		t.Fatalf("expected delete removed prompt error, got %v", err)
	}
}
