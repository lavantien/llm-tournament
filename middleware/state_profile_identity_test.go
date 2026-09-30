package middleware

import (
	"testing"
)

// Prompts reference profiles by profile_id with ON DELETE SET NULL, and the
// ids are AUTOINCREMENT. Rewriting the profile list by delete and reinsert
// would hand every surviving profile a fresh id and null out every prompt
// link. Handlers round-trip profiles read from ReadProfiles back through
// WriteProfiles, which is the flow mirrored here.

func TestWriteProfileSuite_PreservesLinksAndIDsAcrossRewrite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WriteProfileSuite("default", []Profile{{Name: "P1", Description: "d1"}}); err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1", Profile: "P1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	var originalID int
	if err := db.QueryRow("SELECT id FROM profiles WHERE name = 'P1'").Scan(&originalID); err != nil {
		t.Fatalf("failed to read profile id: %v", err)
	}

	// Add a profile the way AddProfile does: read, append, write back.
	profiles := ReadProfiles()
	profiles = append(profiles, Profile{Name: "P2", Description: "d2"})
	if err := WriteProfiles(profiles); err != nil {
		t.Fatalf("WriteProfiles failed: %v", err)
	}

	prompts, err := ReadPromptSuite("default")
	if err != nil {
		t.Fatalf("ReadPromptSuite failed: %v", err)
	}
	if len(prompts) != 1 || prompts[0].Profile != "P1" {
		t.Fatalf("expected prompt to still resolve P1, got %#v", prompts)
	}

	var currentID int
	if err := db.QueryRow("SELECT id FROM profiles WHERE name = 'P1'").Scan(&currentID); err != nil {
		t.Fatalf("failed to read profile id after rewrite: %v", err)
	}
	if currentID != originalID {
		t.Fatalf("expected P1 to keep row id %d across the rewrite, got %d", originalID, currentID)
	}
}

func TestWriteProfileSuite_RenameKeepsRowID(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WriteProfileSuite("default", []Profile{{Name: "OldName", Description: "d"}}); err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}

	var originalID int
	if err := db.QueryRow("SELECT id FROM profiles WHERE name = 'OldName'").Scan(&originalID); err != nil {
		t.Fatalf("failed to read profile id: %v", err)
	}

	// Rename the way EditProfile does: read, rename in place, write back.
	profiles := ReadProfiles()
	profiles[0].Name = "NewName"
	if err := WriteProfiles(profiles); err != nil {
		t.Fatalf("WriteProfiles failed: %v", err)
	}

	var renamedID int
	if err := db.QueryRow("SELECT id FROM profiles WHERE name = 'NewName'").Scan(&renamedID); err != nil {
		t.Fatalf("failed to read renamed profile id: %v", err)
	}
	if renamedID != originalID {
		t.Fatalf("expected rename to keep row id %d, got %d", originalID, renamedID)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM profiles WHERE name = 'OldName'").Scan(&count); err != nil {
		t.Fatalf("failed to count old profile rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected old name to be gone, got %d rows", count)
	}
}

func TestWriteProfileSuite_ResetSeversLinks(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WriteProfileSuite("default", []Profile{{Name: "P1", Description: "d"}}); err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1", Profile: "P1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// ResetProfiles intentionally wipes the profile list.
	if err := WriteProfiles([]Profile{}); err != nil {
		t.Fatalf("WriteProfiles failed: %v", err)
	}

	if profiles := ReadProfiles(); len(profiles) != 0 {
		t.Fatalf("expected reset to remove all profiles, got %#v", profiles)
	}
	prompts, err := ReadPromptSuite("default")
	if err != nil {
		t.Fatalf("ReadPromptSuite failed: %v", err)
	}
	if len(prompts) != 1 || prompts[0].Profile != "" {
		t.Fatalf("expected reset to sever prompt links, got %#v", prompts)
	}
}
