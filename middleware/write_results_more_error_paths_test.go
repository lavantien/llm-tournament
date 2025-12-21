package middleware

import (
	"strings"
	"testing"
)

func TestWriteResults_PrepareScoreInsertError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	suiteID, err := GetSuiteID("default")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
	if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "Model1", suiteID); err != nil {
		t.Fatalf("insert model: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE scores"); err != nil {
		t.Fatalf("drop scores: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_id INTEGER NOT NULL,
			prompt_id INTEGER NOT NULL
		);`); err != nil {
		t.Fatalf("create scores: %v", err)
	}

	err = WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{1}},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to prepare score insert") {
		t.Fatalf("expected prepare score insert error, got %v", err)
	}
}

func TestWriteResults_QueryModelError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE models"); err != nil {
		t.Fatalf("drop models: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE models_data (
			name TEXT NOT NULL,
			suite_id INTEGER NOT NULL
		);`); err != nil {
		t.Fatalf("create models_data: %v", err)
	}
	if _, err := db.Exec("INSERT INTO models_data (name, suite_id) VALUES ('Model1', 1)"); err != nil {
		t.Fatalf("insert models_data: %v", err)
	}
	if _, err := db.Exec(`CREATE VIEW models AS
			SELECT NULL AS id, name, suite_id FROM models_data;`); err != nil {
		t.Fatalf("create models view: %v", err)
	}

	err := WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{1}},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to query model") {
		t.Fatalf("expected query model error, got %v", err)
	}
}
