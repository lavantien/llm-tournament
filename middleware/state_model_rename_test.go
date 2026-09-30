package middleware

import (
	"strings"
	"testing"
)

func TestRenameModel_PreservesRowIDResponsesAndScores(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
	if err := WriteResults("default", map[string]Result{
		"ModelA": {Scores: []int{42}},
	}); err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	var modelID, promptID int
	if err := db.QueryRow("SELECT id FROM models WHERE name = 'ModelA'").Scan(&modelID); err != nil {
		t.Fatalf("failed to read model id: %v", err)
	}
	if err := db.QueryRow("SELECT id FROM prompts WHERE text = 'p1'").Scan(&promptID); err != nil {
		t.Fatalf("failed to read prompt id: %v", err)
	}
	if _, err := db.Exec(
		"INSERT INTO model_responses (model_id, prompt_id, response_text) VALUES (?, ?, ?)",
		modelID, promptID, "saved response",
	); err != nil {
		t.Fatalf("failed to seed model response: %v", err)
	}

	if err := RenameModel("default", "ModelA", "ModelB"); err != nil {
		t.Fatalf("RenameModel failed: %v", err)
	}

	var renamedID int
	if err := db.QueryRow("SELECT id FROM models WHERE name = 'ModelB'").Scan(&renamedID); err != nil {
		t.Fatalf("failed to read renamed model id: %v", err)
	}
	if renamedID != modelID {
		t.Fatalf("expected rename to keep row id %d, got %d", modelID, renamedID)
	}

	var responseText string
	if err := db.QueryRow(
		"SELECT response_text FROM model_responses WHERE model_id = ? AND prompt_id = ?",
		modelID, promptID,
	).Scan(&responseText); err != nil {
		t.Fatalf("expected model response to survive the rename: %v", err)
	}
	if responseText != "saved response" {
		t.Fatalf("expected saved response text, got %q", responseText)
	}

	if scores := ReadResults()["ModelB"].Scores; len(scores) != 1 || scores[0] != 42 {
		t.Fatalf("expected score 42 to survive the rename, got %v", scores)
	}
}

func TestRenameModel_MissingModelIsNoOp(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := RenameModel("default", "Ghost", "Renamed"); err != nil {
		t.Fatalf("expected renaming a missing model to be a no-op, got %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM models").Scan(&count); err != nil {
		t.Fatalf("failed to count models: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no model rows, got %d", count)
	}
}

func TestRenameModel_DuplicateNewName_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := WriteResults("default", map[string]Result{
		"ModelA": {Scores: []int{1}},
		"ModelB": {Scores: []int{2}},
	}); err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	err := RenameModel("default", "ModelA", "ModelB")
	if err == nil {
		t.Fatalf("expected error when renaming to an existing name")
	}
	if !strings.Contains(err.Error(), "failed to rename model") {
		t.Fatalf("expected rename error, got %v", err)
	}
}

func TestRenameModel_GetSuiteIDError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := CloseDB(); err != nil {
		t.Fatalf("CloseDB failed: %v", err)
	}

	err := RenameModel("default", "ModelA", "ModelB")
	if err == nil {
		t.Fatalf("expected error when database is closed")
	}
	if !strings.Contains(err.Error(), "failed to get suite ID") {
		t.Fatalf("expected suite id error, got %v", err)
	}
}

func TestSQLiteDataStore_RenameModel(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}
	if err := ds.RenameModel("default", "ModelA", "ModelB"); err != nil {
		t.Errorf("RenameModel failed: %v", err)
	}
}
