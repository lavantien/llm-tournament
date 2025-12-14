package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"llm-tournament/middleware"

	_ "github.com/mattn/go-sqlite3"
)

// setupProfilesTestDB creates a test database for profile handler tests
func setupProfilesTestDB(t *testing.T) func() {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	err := middleware.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	return func() {
		middleware.CloseDB()
	}
}

func TestAddProfileHandler_Success(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("profile_name", "TestProfile")
	form.Add("profile_description", "Test Description")

	req := httptest.NewRequest("POST", "/add_profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddProfileHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify profile was added
	profiles := middleware.ReadProfiles()
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "TestProfile" {
		t.Errorf("expected profile name 'TestProfile', got %q", profiles[0].Name)
	}
}

func TestAddProfileHandler_EmptyName(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("profile_name", "")
	form.Add("profile_description", "Description")

	req := httptest.NewRequest("POST", "/add_profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddProfileHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestEditProfileHandler_POST_Success(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// First add a profile
	addForm := url.Values{}
	addForm.Add("profile_name", "OldProfileName")
	addForm.Add("profile_description", "Old Description")

	addReq := httptest.NewRequest("POST", "/add_profile", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddProfileHandler(httptest.NewRecorder(), addReq)

	// Edit the profile
	editForm := url.Values{}
	editForm.Add("index", "0")
	editForm.Add("profile_name", "NewProfileName")
	editForm.Add("profile_description", "New Description")

	editReq := httptest.NewRequest("POST", "/edit_profile", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditProfileHandler(editRR, editReq)

	if editRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, editRR.Code)
	}

	// Verify profile was renamed
	profiles := middleware.ReadProfiles()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "NewProfileName" {
		t.Errorf("expected profile name 'NewProfileName', got %q", profiles[0].Name)
	}
}

func TestEditProfileHandler_POST_EmptyName(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// First add a profile
	addForm := url.Values{}
	addForm.Add("profile_name", "ExistingProfile")
	addReq := httptest.NewRequest("POST", "/add_profile", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddProfileHandler(httptest.NewRecorder(), addReq)

	// Try to edit with empty name
	editForm := url.Values{}
	editForm.Add("index", "0")
	editForm.Add("profile_name", "")

	editReq := httptest.NewRequest("POST", "/edit_profile", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditProfileHandler(editRR, editReq)

	if editRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, editRR.Code)
	}
}

func TestEditProfileHandler_POST_InvalidIndex(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	editForm := url.Values{}
	editForm.Add("index", "invalid")
	editForm.Add("profile_name", "NewName")

	editReq := httptest.NewRequest("POST", "/edit_profile", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditProfileHandler(editRR, editReq)

	if editRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, editRR.Code)
	}
}

func TestEditProfileHandler_GET_InvalidIndex(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/edit_profile?index=invalid", nil)
	rr := httptest.NewRecorder()
	EditProfileHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestDeleteProfileHandler_POST_Success(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// First add a profile
	addForm := url.Values{}
	addForm.Add("profile_name", "ProfileToDelete")

	addReq := httptest.NewRequest("POST", "/add_profile", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddProfileHandler(httptest.NewRecorder(), addReq)

	// Delete the profile
	deleteForm := url.Values{}
	deleteForm.Add("index", "0")

	deleteReq := httptest.NewRequest("POST", "/delete_profile", strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	deleteRR := httptest.NewRecorder()
	DeleteProfileHandler(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, deleteRR.Code)
	}

	// Verify profile was deleted
	profiles := middleware.ReadProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestDeleteProfileHandler_POST_InvalidIndex(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	deleteForm := url.Values{}
	deleteForm.Add("index", "not_a_number")

	deleteReq := httptest.NewRequest("POST", "/delete_profile", strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	deleteRR := httptest.NewRecorder()
	DeleteProfileHandler(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, deleteRR.Code)
	}
}

func TestDeleteProfileHandler_GET_InvalidIndex(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/delete_profile?index=invalid", nil)
	rr := httptest.NewRecorder()
	DeleteProfileHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestResetProfilesHandler_POST(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add some profiles first
	addForm := url.Values{}
	addForm.Add("profile_name", "Profile1")

	addReq := httptest.NewRequest("POST", "/add_profile", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddProfileHandler(httptest.NewRecorder(), addReq)

	addForm2 := url.Values{}
	addForm2.Add("profile_name", "Profile2")

	addReq2 := httptest.NewRequest("POST", "/add_profile", strings.NewReader(addForm2.Encode()))
	addReq2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddProfileHandler(httptest.NewRecorder(), addReq2)

	// Verify we have 2 profiles
	profiles := middleware.ReadProfiles()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	// Reset profiles
	resetReq := httptest.NewRequest("POST", "/reset_profiles", nil)
	resetRR := httptest.NewRecorder()
	ResetProfilesHandler(resetRR, resetReq)

	if resetRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, resetRR.Code)
	}

	// Verify profiles were reset
	profiles = middleware.ReadProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles after reset, got %d", len(profiles))
	}
}

func TestDeleteProfileHandler_POST_OutOfRange(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add one profile
	addForm := url.Values{}
	addForm.Add("profile_name", "OnlyProfile")

	addReq := httptest.NewRequest("POST", "/add_profile", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddProfileHandler(httptest.NewRecorder(), addReq)

	// Try to delete with out-of-range index
	deleteForm := url.Values{}
	deleteForm.Add("index", "99")

	deleteReq := httptest.NewRequest("POST", "/delete_profile", strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	deleteRR := httptest.NewRecorder()
	DeleteProfileHandler(deleteRR, deleteReq)

	// Should redirect (the handler doesn't explicitly error on out-of-range)
	// The profile should still exist because the index was out of range
	profiles := middleware.ReadProfiles()
	if len(profiles) != 1 {
		t.Errorf("expected profile to still exist, got %d profiles", len(profiles))
	}
}
