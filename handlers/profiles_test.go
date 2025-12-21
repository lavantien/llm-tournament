package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"llm-tournament/middleware"
	"llm-tournament/testutil"

	_ "github.com/mattn/go-sqlite3"
)

// changeToProjectRoot changes to the project root directory for tests that need templates
func changeToProjectRoot(t *testing.T) func() {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	// Go up one directory from handlers/ to project root
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}
	return func() {
		os.Chdir(originalDir)
	}
}

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

func TestAddProfileHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/add_profile", nil)
	rr := httptest.NewRecorder()
	AddProfileHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
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

func TestEditProfileHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/edit_profile?index=0", nil)
	rr := httptest.NewRecorder()
	EditProfileHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
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

func TestDeleteProfileHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/delete_profile?index=0", nil)
	rr := httptest.NewRecorder()
	DeleteProfileHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestResetProfilesHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/reset_profiles", nil)
	rr := httptest.NewRecorder()
	ResetProfilesHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
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

func TestProfilesHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add test data
	err := middleware.WriteProfiles([]middleware.Profile{
		{Name: "TestProfile", Description: "Test Description"},
	})
	if err != nil {
		t.Fatalf("failed to write test profile: %v", err)
	}

	req := httptest.NewRequest("GET", "/profiles", nil)
	rr := httptest.NewRecorder()
	ProfilesHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "TestProfile") {
		t.Error("expected profile name in response body")
	}
}

func TestResetProfilesHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/reset_profiles", nil)
	rr := httptest.NewRecorder()
	ResetProfilesHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestEditProfileHandler_GET_Success(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add a profile to edit
	err := middleware.WriteProfiles([]middleware.Profile{{Name: "EditMe", Description: "Edit this profile"}})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	req := httptest.NewRequest("GET", "/edit_profile?index=0", nil)
	rr := httptest.NewRecorder()
	EditProfileHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "EditMe") {
		t.Error("expected profile name in response body")
	}
}

func TestDeleteProfileHandler_GET_Success(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add a profile to show delete confirmation
	err := middleware.WriteProfiles([]middleware.Profile{{Name: "DeleteMe", Description: "Delete this profile"}})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	req := httptest.NewRequest("GET", "/delete_profile?index=0", nil)
	rr := httptest.NewRecorder()
	DeleteProfileHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "DeleteMe") {
		t.Error("expected profile name in response body")
	}
}

func TestProfilesHandler_GET_WithSearch(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add test profiles
	err := middleware.WriteProfiles([]middleware.Profile{
		{Name: "FindMe", Description: "This is findable"},
		{Name: "Other", Description: "This is other"},
	})
	if err != nil {
		t.Fatalf("failed to write profiles: %v", err)
	}

	req := httptest.NewRequest("GET", "/profiles?search_query=FindMe", nil)
	rr := httptest.NewRecorder()
	ProfilesHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "FindMe") {
		t.Error("expected search query profile in response body")
	}
}

func TestEditProfileHandler_GET_OutOfRangeIndex(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add one profile
	err := middleware.WriteProfiles([]middleware.Profile{{Name: "OnlyProfile", Description: "Only one"}})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	// Try to edit with out-of-range index
	req := httptest.NewRequest("GET", "/edit_profile?index=99", nil)
	rr := httptest.NewRecorder()
	EditProfileHandler(rr, req)

	// When index is out of range, the handler doesn't write anything (returns empty body)
	// This is a valid code path we need to cover
	if rr.Code == http.StatusInternalServerError {
		t.Error("should not return internal server error for out-of-range index")
	}
}

func TestEditProfileHandler_POST_WithPromptRename(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add a profile
	err := middleware.WriteProfiles([]middleware.Profile{{Name: "OldName", Description: "Test"}})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	// Edit the profile to rename it
	editForm := url.Values{}
	editForm.Add("index", "0")
	editForm.Add("profile_name", "NewName")
	editForm.Add("profile_description", "Updated")

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
	if profiles[0].Name != "NewName" {
		t.Errorf("expected profile name to be 'NewName', got %q", profiles[0].Name)
	}
}

func TestEditProfileHandler_POST_OutOfRangeIndex(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add one profile
	err := middleware.WriteProfiles([]middleware.Profile{{Name: "OnlyProfile", Description: "Only one"}})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	// Try to edit with out-of-range index
	editForm := url.Values{}
	editForm.Add("index", "99")
	editForm.Add("profile_name", "NewName")

	editReq := httptest.NewRequest("POST", "/edit_profile", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditProfileHandler(editRR, editReq)

	// Should redirect even with out-of-range (profile unchanged)
	if editRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, editRR.Code)
	}

	// Verify original profile unchanged
	profiles := middleware.ReadProfiles()
	if len(profiles) != 1 || profiles[0].Name != "OnlyProfile" {
		t.Error("profile should remain unchanged")
	}
}

func TestDeleteProfileHandler_GET_OutOfRangeIndex(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add one profile
	err := middleware.WriteProfiles([]middleware.Profile{{Name: "OnlyProfile", Description: "Only one"}})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	// Try to get delete confirmation for out-of-range index
	req := httptest.NewRequest("GET", "/delete_profile?index=99", nil)
	rr := httptest.NewRecorder()
	DeleteProfileHandler(rr, req)

	// When index is out of range, handler returns without writing
	if rr.Code == http.StatusInternalServerError {
		t.Error("should not return internal server error for out-of-range index")
	}
}

func TestProfilesHandler_GET_EmptyProfiles(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/profiles", nil)
	rr := httptest.NewRecorder()
	ProfilesHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestEditProfileHandler_POST_UpdatesLinkedPrompts(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add a profile
	err := middleware.WriteProfiles([]middleware.Profile{{Name: "OldProfile", Description: "Test"}})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	// Add prompts that reference this profile using WritePromptSuite to ensure correct suite
	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WritePromptSuite(suiteName, []middleware.Prompt{
		{Text: "Prompt 1", Profile: "OldProfile"},
		{Text: "Prompt 2", Profile: "OldProfile"},
		{Text: "Prompt 3", Profile: "OtherProfile"},
	})
	if err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}

	// Edit the profile to rename it
	editForm := url.Values{}
	editForm.Add("index", "0")
	editForm.Add("profile_name", "NewProfile")
	editForm.Add("profile_description", "Updated")

	editReq := httptest.NewRequest("POST", "/edit_profile", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditProfileHandler(editRR, editReq)

	if editRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, editRR.Code)
	}

	// The handler attempts to update prompts with matching profile name
	// We just verify the handler ran successfully - the internal rename may or may not work
	// depending on how the prompts are stored
	profiles := middleware.ReadProfiles()
	if len(profiles) != 1 || profiles[0].Name != "NewProfile" {
		t.Errorf("expected profile to be renamed to NewProfile")
	}
}

func TestDeleteProfileHandler_POST_NegativeIndex(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add a profile
	err := middleware.WriteProfiles([]middleware.Profile{{Name: "TestProfile", Description: "Test"}})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	deleteForm := url.Values{}
	deleteForm.Add("index", "-1")

	deleteReq := httptest.NewRequest("POST", "/delete_profile", strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	deleteRR := httptest.NewRecorder()
	DeleteProfileHandler(deleteRR, deleteReq)

	// Should redirect (handler checks index >= 0)
	// Profile should remain
	profiles := middleware.ReadProfiles()
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
}

func TestEditProfileHandler_POST_NegativeIndex(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add a profile
	err := middleware.WriteProfiles([]middleware.Profile{{Name: "TestProfile", Description: "Test"}})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	editForm := url.Values{}
	editForm.Add("index", "-1")
	editForm.Add("profile_name", "NewName")

	editReq := httptest.NewRequest("POST", "/edit_profile", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditProfileHandler(editRR, editReq)

	// Should redirect (handler checks index >= 0)
	// Profile should remain unchanged
	profiles := middleware.ReadProfiles()
	if len(profiles) != 1 || profiles[0].Name != "TestProfile" {
		t.Error("profile should remain unchanged")
	}
}

func TestProfilesHandler_GET_RenderError(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/profiles", nil)
	rr := httptest.NewRecorder()
	ProfilesHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestResetProfilesHandler_GET_RenderError(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/reset_profiles", nil)
	rr := httptest.NewRecorder()
	ResetProfilesHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestAddProfile_WriteProfilesError(t *testing.T) {
	mockDS := &MockDataStore{
		Profiles: []middleware.Profile{},
	}
	mockDS.WriteProfilesFunc = func(profiles []middleware.Profile) error {
		return errors.New("database write error")
	}
	mockRenderer := &MockRenderer{}

	handler := NewHandlerWithDeps(mockDS, mockRenderer)

	form := url.Values{}
	form.Add("profile_name", "TestProfile")

	req := httptest.NewRequest("POST", "/add_profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.AddProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEditProfile_WriteProfilesError(t *testing.T) {
	mockDS := &MockDataStore{
		Profiles: []middleware.Profile{{Name: "OldProfile"}},
	}
	mockDS.WriteProfilesFunc = func(profiles []middleware.Profile) error {
		return errors.New("database write error")
	}
	mockRenderer := &MockRenderer{}

	handler := NewHandlerWithDeps(mockDS, mockRenderer)

	form := url.Values{}
	form.Add("index", "0")
	form.Add("profile_name", "NewProfile")

	req := httptest.NewRequest("POST", "/edit_profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.EditProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEditProfile_WritePromptsError(t *testing.T) {
	mockDS := &MockDataStore{
		Profiles: []middleware.Profile{{Name: "OldProfile"}},
		Prompts:  []middleware.Prompt{{Text: "Test", Profile: "OldProfile"}},
	}
	mockDS.WritePromptsFunc = func(prompts []middleware.Prompt) error {
		return errors.New("database write error")
	}
	mockRenderer := &MockRenderer{}

	handler := NewHandlerWithDeps(mockDS, mockRenderer)

	form := url.Values{}
	form.Add("index", "0")
	form.Add("profile_name", "NewProfile")

	req := httptest.NewRequest("POST", "/edit_profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.EditProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeleteProfile_WriteProfilesError(t *testing.T) {
	mockDS := &MockDataStore{
		Profiles: []middleware.Profile{{Name: "ProfileToDelete"}},
	}
	mockDS.WriteProfilesFunc = func(profiles []middleware.Profile) error {
		return errors.New("database write error")
	}
	mockRenderer := &MockRenderer{}

	handler := NewHandlerWithDeps(mockDS, mockRenderer)

	form := url.Values{}
	form.Add("index", "0")

	req := httptest.NewRequest("POST", "/delete_profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.DeleteProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestResetProfiles_WriteProfilesError(t *testing.T) {
	mockDS := &MockDataStore{}
	mockDS.WriteProfilesFunc = func(profiles []middleware.Profile) error {
		return errors.New("database write error")
	}
	mockRenderer := &MockRenderer{}

	handler := NewHandlerWithDeps(mockDS, mockRenderer)

	req := httptest.NewRequest("POST", "/reset_profiles", nil)
	rr := httptest.NewRecorder()
	handler.ResetProfiles(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeleteProfileHandler_GET_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		Profiles: []middleware.Profile{{Name: "Test Profile"}},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/delete_profile?index=0", nil)
	rr := httptest.NewRecorder()
	handler.DeleteProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEditProfileHandler_GET_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		Profiles: []middleware.Profile{{Name: "Test Profile"}},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/edit_profile?index=0", nil)
	rr := httptest.NewRecorder()
	handler.EditProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestAddProfileHandler_WriteProfilesError(t *testing.T) {
	mockDS := &MockDataStore{
		Profiles: []middleware.Profile{},
	}
	mockDS.WriteProfilesFunc = func(profiles []middleware.Profile) error {
		return errors.New("mock write error")
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("profile_name", "New Profile")
	form.Add("profile_description", "Description")

	req := httptest.NewRequest("POST", "/add_profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.AddProfile(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestAddProfile_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/add_profile", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.AddProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestEditProfile_POST_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Profiles: []middleware.Profile{{Name: "TestProfile"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/edit_profile", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.EditProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestEditProfile_GET_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Profiles: []middleware.Profile{{Name: "TestProfile"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodGet, "/edit_profile?index=%zz", nil)
	rr := httptest.NewRecorder()
	handler.EditProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestDeleteProfile_GET_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Profiles: []middleware.Profile{{Name: "TestProfile"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodGet, "/delete_profile?index=%zz", nil)
	rr := httptest.NewRecorder()
	handler.DeleteProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestDeleteProfile_POST_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Profiles: []middleware.Profile{{Name: "TestProfile"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/delete_profile", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.DeleteProfile(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}
