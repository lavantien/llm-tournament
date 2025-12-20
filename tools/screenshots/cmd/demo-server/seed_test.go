package main

import (
	"os"
	"path/filepath"
	"strings"
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
	if promptCount < 12 {
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

func TestEnsureDemoEncryptionKey_DoesNotOverrideExistingValue(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "already-set")
	ensureDemoEncryptionKey()
	if got := os.Getenv("ENCRYPTION_KEY"); got != "already-set" {
		t.Fatalf("expected ENCRYPTION_KEY to be preserved, got %q", got)
	}
}

func TestSeedDemoData_ReturnsError_WhenDBClosed(t *testing.T) {
	t.Helper()
	ensureDemoEncryptionKey()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "demo.db")
	if err := middleware.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	if err := middleware.CloseDB(); err != nil {
		t.Fatalf("CloseDB: %v", err)
	}

	if err := seedDemoData(); err == nil {
		t.Fatalf("expected error when database is closed")
	}
}

func TestSeedDemoData_ReturnsError_WhenWriteProfilesFails(t *testing.T) {
	t.Helper()
	ensureDemoEncryptionKey()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "demo.db")
	if err := middleware.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = middleware.CloseDB() })

	if _, err := middleware.GetDB().Exec("DROP TABLE profiles"); err != nil {
		t.Fatalf("drop profiles: %v", err)
	}

	err := seedDemoData()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "write profiles") {
		t.Fatalf("expected write profiles error, got %v", err)
	}
}

func TestSeedDemoData_ReturnsError_WhenWritePromptsFails(t *testing.T) {
	t.Helper()
	ensureDemoEncryptionKey()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "demo.db")
	if err := middleware.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = middleware.CloseDB() })

	if _, err := middleware.GetDB().Exec("DROP TABLE prompts"); err != nil {
		t.Fatalf("drop prompts: %v", err)
	}

	err := seedDemoData()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "write prompts") {
		t.Fatalf("expected write prompts error, got %v", err)
	}
}

func TestSeedDemoData_ReturnsError_WhenWriteResultsFails(t *testing.T) {
	t.Helper()
	ensureDemoEncryptionKey()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "demo.db")
	if err := middleware.InitDB(dbPath); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = middleware.CloseDB() })

	if _, err := middleware.GetDB().Exec("DROP TABLE scores"); err != nil {
		t.Fatalf("drop scores: %v", err)
	}

	err := seedDemoData()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "write results") {
		t.Fatalf("expected write results error, got %v", err)
	}
}
