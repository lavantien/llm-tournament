package middleware

import (
	"strings"
	"testing"
)

func TestWriteProfileSuite_PrepareError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE profiles"); err != nil {
		t.Fatalf("drop profiles: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			suite_id INTEGER NOT NULL
		);`); err != nil {
		t.Fatalf("create profiles: %v", err)
	}

	err := WriteProfileSuite("default", []Profile{{Name: "p1", Description: "desc"}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to prepare profile insert") {
		t.Fatalf("expected prepare error, got %v", err)
	}
}

func TestWritePromptSuite_PrepareError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}
	if _, err := db.Exec("DROP TABLE prompts"); err != nil {
		t.Fatalf("drop prompts: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			solution TEXT DEFAULT '',
			suite_id INTEGER NOT NULL,
			display_order INTEGER NOT NULL
		);`); err != nil {
		t.Fatalf("create prompts: %v", err)
	}

	err := WritePromptSuite("default", []Prompt{{Text: "p1"}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to prepare prompt insert") {
		t.Fatalf("expected prepare error, got %v", err)
	}
}
