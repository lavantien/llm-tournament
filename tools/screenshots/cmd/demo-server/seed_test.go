package main

import (
	"os"
	"path/filepath"
	"testing"

	"llm-tournament/middleware"
)

func TestSeedDemoData_CreatesCoreRecords(t *testing.T) {
	t.Helper()

	ensureDemoEncryptionKey()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "demo.db")

	if err := middleware.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = middleware.CloseDB() })

	if err := seedDemoData(); err != nil {
		t.Fatalf("seedDemoData: %v", err)
	}

	db := middleware.GetDB()

	var promptCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&promptCount); err != nil {
		t.Fatalf("count prompts: %v", err)
	}
	if promptCount < 3 {
		t.Fatalf("expected prompts seeded, got %d", promptCount)
	}

	var modelCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM models").Scan(&modelCount); err != nil {
		t.Fatalf("count models: %v", err)
	}
	if modelCount < 3 {
		t.Fatalf("expected models seeded, got %d", modelCount)
	}

	var scoreCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM scores").Scan(&scoreCount); err != nil {
		t.Fatalf("count scores: %v", err)
	}
	if scoreCount < 3 {
		t.Fatalf("expected scores seeded, got %d", scoreCount)
	}

	if v := os.Getenv("ENCRYPTION_KEY"); v == "" {
		t.Fatalf("expected ENCRYPTION_KEY to be set for demo seeding")
	}
}
