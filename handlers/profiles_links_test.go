package handlers

import (
	"errors"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

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

func TestEditProfileHandler_POST_UpdatesLinkedPrompts(t *testing.T) {
	cleanup := setupProfilesTestDB(t)
	defer cleanup()

	// Add the profile to rename plus a second profile as an untouched control
	err := middleware.WriteProfiles([]middleware.Profile{
		{Name: "OldProfile", Description: "Test"},
		{Name: "OtherProfile", Description: "Control"},
	})
	if err != nil {
		t.Fatalf("failed to write profiles: %v", err)
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
	profiles := middleware.ReadProfiles()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	renamed := false
	for _, p := range profiles {
		if p.Name == "NewProfile" {
			renamed = true
		}
	}
	if !renamed {
		t.Error("expected profile to be renamed to NewProfile")
	}

	// Prompts linked to the renamed profile must carry the new name,
	// while prompts of other profiles stay untouched
	prompts := middleware.ReadPrompts()
	if len(prompts) != 3 {
		t.Fatalf("expected 3 prompts after rename, got %d", len(prompts))
	}
	for _, p := range prompts {
		switch p.Text {
		case "Prompt 1", "Prompt 2":
			if p.Profile != "NewProfile" {
				t.Errorf("expected %s profile rewritten to NewProfile, got %q", p.Text, p.Profile)
			}
		case "Prompt 3":
			if p.Profile != "OtherProfile" {
				t.Errorf("expected Prompt 3 profile to stay OtherProfile, got %q", p.Profile)
			}
		}
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
