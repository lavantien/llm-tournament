package middleware

import (
	"database/sql"
	"strings"
	"testing"
)

// With the suites table empty there is no current suite and no default row,
// so GetCurrentSuiteID must still terminate: the old implementation recursed
// into the identical state after a zero-row UPDATE.
func TestGetCurrentSuiteID_EmptySuitesTableTerminates(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if _, err := db.Exec("DELETE FROM suites"); err != nil {
		t.Fatalf("failed to empty suites table: %v", err)
	}

	suiteID, err := GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("expected GetCurrentSuiteID to terminate without error, got %v", err)
	}
	if suiteID <= 0 {
		t.Fatalf("expected a valid suite ID, got %d", suiteID)
	}
	if !SuiteExists("default") {
		t.Error("expected the default suite to be recreated")
	}
	var isCurrent bool
	if err := db.QueryRow("SELECT is_current FROM suites WHERE name = 'default'").Scan(&isCurrent); err != nil {
		t.Fatalf("failed to query default suite: %v", err)
	}
	if !isCurrent {
		t.Error("expected the recreated default suite to be current")
	}
}

func TestGetCurrentSuiteID_InsertDefaultError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if _, err := db.Exec("DELETE FROM suites"); err != nil {
		t.Fatalf("failed to empty suites table: %v", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER abort_current_suite_insert
		BEFORE INSERT ON suites
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("failed to create trigger: %v", err)
	}

	_, err := GetCurrentSuiteID()
	if err == nil {
		t.Fatal("expected error when creating the default suite fails")
	}
	if !strings.Contains(err.Error(), "failed to create default suite") {
		t.Fatalf("expected create default suite error, got %v", err)
	}
}

// countCurrentSuites reports how many suites are flagged current.
func countCurrentSuites(t *testing.T) int {
	t.Helper()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM suites WHERE is_current = 1").Scan(&count); err != nil {
		t.Fatalf("failed to count current suites: %v", err)
	}
	return count
}

// The PRAGMA foreign_keys statement is per-connection, but database/sql
// pools connections, so the setting must travel in the DSN where every
// pooled connection picks it up.
func TestInitDB_DSNEnablesForeignKeys(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	var gotDSN string
	original := sqlOpen
	sqlOpen = func(driver, dsn string) (*sql.DB, error) {
		gotDSN = dsn
		return original(driver, dsn)
	}
	t.Cleanup(func() { sqlOpen = original })

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if !strings.Contains(gotDSN, "_foreign_keys=on") {
		t.Fatalf("expected DSN to enable foreign keys per connection, got %q", gotDSN)
	}
}

// Foreign-key enforcement must hold on connections opened after InitDB's
// first one, where a one-shot PRAGMA never ran.
func TestInitDB_ForeignKeysHoldOnFreshPooledConnections(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	suiteID, err := GetSuiteID("default")
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}
	if _, err := db.Exec("INSERT INTO profiles (name, suite_id) VALUES ('fk-profile', ?)", suiteID); err != nil {
		t.Fatalf("failed to insert profile: %v", err)
	}
	var profileID int
	if err := db.QueryRow("SELECT id FROM profiles WHERE name = 'fk-profile'").Scan(&profileID); err != nil {
		t.Fatalf("failed to query profile id: %v", err)
	}
	if _, err := db.Exec("INSERT INTO prompts (text, suite_id, profile_id, display_order) VALUES ('fk-prompt', ?, ?, 0)", suiteID, profileID); err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}

	// Hold one pooled connection busy so the delete below runs on a
	// freshly opened connection.
	held, err := db.Query("SELECT id FROM suites")
	if err != nil {
		t.Fatalf("failed to hold a pooled connection: %v", err)
	}
	defer func() { _ = held.Close() }()

	if _, err := db.Exec("DELETE FROM profiles WHERE id = ?", profileID); err != nil {
		t.Fatalf("failed to delete profile: %v", err)
	}

	// ON DELETE SET NULL must have fired on the fresh connection.
	var nulls int
	if err := db.QueryRow("SELECT COUNT(*) FROM prompts WHERE profile_id IS NULL").Scan(&nulls); err != nil {
		t.Fatalf("failed to count null profile references: %v", err)
	}
	if nulls != 1 {
		t.Fatalf("expected 1 nulled profile reference after cascade on a fresh pooled connection, got %d", nulls)
	}
}

// A concurrent SetCurrentSuite commit can land between the ErrNoRows read
// and the recovery writes; the trigger emulates exactly that interleaving.
// Recovery must end with a single current row, the default suite.
func TestGetCurrentSuiteID_RecoveryIsExclusive(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if _, err := db.Exec("DELETE FROM suites"); err != nil {
		t.Fatalf("failed to empty suites table: %v", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER race_make_other_current
		BEFORE INSERT ON suites
		BEGIN
			INSERT INTO suites (name, is_current) VALUES ('race-x', 1);
		END;`); err != nil {
		t.Fatalf("failed to create trigger: %v", err)
	}

	suiteID, err := GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("GetCurrentSuiteID failed: %v", err)
	}

	if got := countCurrentSuites(t); got != 1 {
		t.Fatalf("expected exactly 1 current suite after recovery, got %d", got)
	}
	var name string
	var id int
	if err := db.QueryRow("SELECT id, name FROM suites WHERE is_current = 1").Scan(&id, &name); err != nil {
		t.Fatalf("failed to query current suite: %v", err)
	}
	if name != "default" {
		t.Errorf("expected default to be the current suite, got %q", name)
	}
	if suiteID != id {
		t.Errorf("returned ID %d, want %d", suiteID, id)
	}
}

// On an emptied suites table the name resolver must run the same recovery as
// the ID resolver, so the default suite actually exists afterwards instead
// of the two resolvers disagreeing.
func TestGetCurrentSuiteName_EmptySuitesTableAgreesWithIDPath(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if _, err := db.Exec("DELETE FROM suites"); err != nil {
		t.Fatalf("failed to empty suites table: %v", err)
	}

	name := GetCurrentSuiteName()
	if name != "default" {
		t.Fatalf("expected 'default', got %q", name)
	}
	if !SuiteExists("default") {
		t.Fatal("expected the default suite to be created by the name recovery")
	}
	if got := countCurrentSuites(t); got != 1 {
		t.Fatalf("expected exactly 1 current suite after name recovery, got %d", got)
	}

	id, err := GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("GetCurrentSuiteID failed after name recovery: %v", err)
	}
	var wantID int
	if err := db.QueryRow("SELECT id FROM suites WHERE name = 'default'").Scan(&wantID); err != nil {
		t.Fatalf("failed to query default suite: %v", err)
	}
	if id != wantID {
		t.Errorf("name and id resolvers disagree: id %d, want %d", id, wantID)
	}
}
