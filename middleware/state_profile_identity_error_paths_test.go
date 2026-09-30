package middleware

import (
	"strings"
	"testing"
)

func TestWriteProfileSuite_UpdateError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := WriteProfileSuite("default", []Profile{{Name: "P1"}, {Name: "P2"}}); err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER abort_profile_update
		BEFORE UPDATE ON profiles
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	profiles := ReadProfiles()
	err := WriteProfiles(profiles)
	if err == nil {
		t.Fatalf("expected error when profile update fails")
	}
	if !strings.Contains(err.Error(), "failed to update profile") {
		t.Fatalf("expected update profile error, got %v", err)
	}
}

func TestWriteProfileSuite_DeleteRemovedError_ReturnsError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := WriteProfileSuite("default", []Profile{{Name: "P1"}, {Name: "P2"}}); err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER abort_profile_delete
		BEFORE DELETE ON profiles
		BEGIN
			SELECT RAISE(FAIL, 'nope');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	profiles := ReadProfiles()
	err := WriteProfiles(profiles[:1])
	if err == nil {
		t.Fatalf("expected error when removed profile deletion fails")
	}
	if !strings.Contains(err.Error(), "failed to delete removed profile") {
		t.Fatalf("expected delete removed profile error, got %v", err)
	}
}
