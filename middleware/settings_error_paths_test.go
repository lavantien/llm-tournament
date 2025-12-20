package middleware

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"llm-tournament/testutil"
)

func TestGetAllSettings_QueryError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("DROP TABLE settings"); err != nil {
		t.Fatalf("DROP TABLE settings failed: %v", err)
	}

	_, err := GetAllSettings()
	if err == nil {
		t.Fatalf("expected error after dropping settings table")
	}
}

func TestGetAllSettings_ScanError(t *testing.T) {
	original := db

	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	t.Cleanup(func() {
		_ = testDB.Close()
		db = original
	})
	db = testDB

	if _, err := testDB.Exec("CREATE TABLE settings (key TEXT, value TEXT)"); err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}
	if _, err := testDB.Exec("INSERT INTO settings (key, value) VALUES ('k', NULL)"); err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	_, err = GetAllSettings()
	if err == nil {
		t.Fatalf("expected scan error for NULL value")
	}
}

func TestGetMaskedAPIKeys_QueryError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if _, err := db.Exec("DROP TABLE settings"); err != nil {
		t.Fatalf("DROP TABLE settings failed: %v", err)
	}

	_, err := GetMaskedAPIKeys()
	if err == nil {
		t.Fatalf("expected error after dropping settings table")
	}
}

func TestGetMaskedAPIKeys_ScanError(t *testing.T) {
	original := db

	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	t.Cleanup(func() {
		_ = testDB.Close()
		db = original
	})
	db = testDB

	if _, err := testDB.Exec("CREATE TABLE settings (key TEXT, value TEXT)"); err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}
	if _, err := testDB.Exec("INSERT INTO settings (key, value) VALUES ('api_key_openai', NULL)"); err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}

	_, err = GetMaskedAPIKeys()
	if err == nil {
		t.Fatalf("expected scan error for NULL api key value")
	}
}

func TestGetMaskedAPIKeys_DecryptErrorMasksAsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	encCleanup := testutil.SetupEncryptionKey(t)
	defer encCleanup()

	// Store an invalid ciphertext for one provider so DecryptAPIKey fails.
	if err := SetSetting("api_key_openai", "not-valid-base64!!!"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	masked, err := GetMaskedAPIKeys()
	if err != nil {
		t.Fatalf("GetMaskedAPIKeys failed: %v", err)
	}

	if got := masked["api_key_openai"]; got != "***ERROR***" {
		t.Fatalf("expected masked value to be ***ERROR***, got %q", got)
	}
}

