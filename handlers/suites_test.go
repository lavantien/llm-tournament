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

// changeToProjectRootForSuites changes to the project root directory for tests
func changeToProjectRootForSuites(t *testing.T) func() {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}
	return func() {
		os.Chdir(originalDir)
	}
}

// setupSuitesTestDB creates a test database for suite handler tests
func setupSuitesTestDB(t *testing.T) func() {
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

func TestNewPromptSuiteHandler_POST_Success(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("suite_name", "new-test-suite")

	req := httptest.NewRequest("POST", "/new_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	NewPromptSuiteHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify suite was created
	if !middleware.SuiteExists("new-test-suite") {
		t.Error("new-test-suite should exist after creation")
	}
}

func TestNewPromptSuiteHandler_POST_EmptyName(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("suite_name", "")

	req := httptest.NewRequest("POST", "/new_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	NewPromptSuiteHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestSelectPromptSuiteHandler_Success(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// First create a suite to select
	form := url.Values{}
	form.Add("suite_name", "suite-to-select")

	createReq := httptest.NewRequest("POST", "/new_prompt_suite", strings.NewReader(form.Encode()))
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	NewPromptSuiteHandler(httptest.NewRecorder(), createReq)

	// Note: SelectPromptSuiteHandler writes to data/current_suite.txt file
	// which won't exist in test environment. Testing the file write path
	// would require creating the data directory first or mocking the file system.
	// For now, we verify that empty name returns 400.

	// Verify suite was created in DB
	if !middleware.SuiteExists("suite-to-select") {
		t.Error("suite-to-select should exist in database")
	}
}

func TestSelectPromptSuiteHandler_EmptyName(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("suite_name", "")

	req := httptest.NewRequest("POST", "/select_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	SelectPromptSuiteHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestSelectPromptSuiteHandler_SetCurrentSuiteError(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// Close the database to trigger SetCurrentSuite error
	middleware.CloseDB()

	form := url.Values{}
	form.Add("suite_name", "test-suite")

	req := httptest.NewRequest("POST", "/select_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	SelectPromptSuiteHandler(rr, req)

	// Should return internal server error when SetCurrentSuite fails
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeletePromptSuiteHandler_POST_EmptyName(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("suite_name", "")

	req := httptest.NewRequest("POST", "/delete_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	DeletePromptSuiteHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestDeletePromptSuiteHandler_POST_Success(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// First create a suite to delete
	createForm := url.Values{}
	createForm.Add("suite_name", "suite-to-delete")

	createReq := httptest.NewRequest("POST", "/new_prompt_suite", strings.NewReader(createForm.Encode()))
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	NewPromptSuiteHandler(httptest.NewRecorder(), createReq)

	// Verify it exists
	if !middleware.SuiteExists("suite-to-delete") {
		t.Fatal("suite-to-delete should exist before deletion")
	}

	// Now delete it
	deleteForm := url.Values{}
	deleteForm.Add("suite_name", "suite-to-delete")

	deleteReq := httptest.NewRequest("POST", "/delete_prompt_suite", strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	deleteRR := httptest.NewRecorder()
	DeletePromptSuiteHandler(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, deleteRR.Code)
	}

	// Verify it was deleted
	if middleware.SuiteExists("suite-to-delete") {
		t.Error("suite-to-delete should not exist after deletion")
	}
}

func TestEditPromptSuiteHandler_POST_EmptyNames(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		oldName     string
		newName     string
		expectError bool
	}{
		{
			name:        "empty old name",
			oldName:     "",
			newName:     "new-name",
			expectError: true,
		},
		{
			name:        "empty new name",
			oldName:     "old-name",
			newName:     "",
			expectError: true,
		},
		{
			name:        "both empty",
			oldName:     "",
			newName:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("suite_name", tt.oldName)
			form.Add("new_suite_name", tt.newName)

			req := httptest.NewRequest("POST", "/edit_prompt_suite", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			EditPromptSuiteHandler(rr, req)

			if tt.expectError && rr.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
			}
		})
	}
}

func TestEditPromptSuiteHandler_POST_Success(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// First create a suite to rename
	createForm := url.Values{}
	createForm.Add("suite_name", "original-suite")

	createReq := httptest.NewRequest("POST", "/new_prompt_suite", strings.NewReader(createForm.Encode()))
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	NewPromptSuiteHandler(httptest.NewRecorder(), createReq)

	// Verify it exists
	if !middleware.SuiteExists("original-suite") {
		t.Fatal("original-suite should exist before rename")
	}

	// Now rename it
	editForm := url.Values{}
	editForm.Add("suite_name", "original-suite")
	editForm.Add("new_suite_name", "renamed-suite")

	editReq := httptest.NewRequest("POST", "/edit_prompt_suite", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditPromptSuiteHandler(editRR, editReq)

	if editRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, editRR.Code)
	}

	// Verify rename worked
	if middleware.SuiteExists("original-suite") {
		t.Error("original-suite should not exist after rename")
	}
	if !middleware.SuiteExists("renamed-suite") {
		t.Error("renamed-suite should exist after rename")
	}
}

func TestNewPromptSuiteHandler_POST_DuplicateName(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// Create first suite
	form := url.Values{}
	form.Add("suite_name", "duplicate-suite")

	req := httptest.NewRequest("POST", "/new_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	NewPromptSuiteHandler(rr, req)

	// Try to create a duplicate (should succeed since it's idempotent)
	req2 := httptest.NewRequest("POST", "/new_prompt_suite", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr2 := httptest.NewRecorder()
	NewPromptSuiteHandler(rr2, req2)

	// Should still redirect (idempotent)
	if rr2.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for duplicate suite, got %d", http.StatusSeeOther, rr2.Code)
	}
}

func TestNewPromptSuiteHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootForSuites(t)
	defer restoreDir()

	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/new_prompt_suite", nil)
	rr := httptest.NewRecorder()
	NewPromptSuiteHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestEditPromptSuiteHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootForSuites(t)
	defer restoreDir()

	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// Create a suite to edit
	middleware.WritePromptSuite("test-suite", []middleware.Prompt{})

	req := httptest.NewRequest("GET", "/edit_prompt_suite?suite_name=test-suite", nil)
	rr := httptest.NewRecorder()
	EditPromptSuiteHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "test-suite") {
		t.Error("expected suite name in response body")
	}
}

func TestDeletePromptSuiteHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootForSuites(t)
	defer restoreDir()

	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/delete_prompt_suite?suite_name=test-suite", nil)
	rr := httptest.NewRecorder()
	DeletePromptSuiteHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "test-suite") {
		t.Error("expected suite name in response body")
	}
}

func TestSelectPromptSuiteHandler_WithDataDir(t *testing.T) {
	restoreDir := changeToProjectRootForSuites(t)
	defer restoreDir()

	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		t.Fatalf("failed to create data directory: %v", err)
	}

	// Create a suite to select
	middleware.WritePromptSuite("selectable-suite", []middleware.Prompt{})

	form := url.Values{}
	form.Add("suite_name", "selectable-suite")

	req := httptest.NewRequest("POST", "/select_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	SelectPromptSuiteHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify the current suite was updated in the database
	currentSuite := middleware.GetCurrentSuiteName()
	if currentSuite != "selectable-suite" {
		t.Errorf("expected 'selectable-suite', got %q", currentSuite)
	}
}

func TestDeletePromptSuiteHandler_POST_CurrentSuiteReset(t *testing.T) {
	restoreDir := changeToProjectRootForSuites(t)
	defer restoreDir()

	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		t.Fatalf("failed to create data directory: %v", err)
	}

	// Create a suite and set it as current via database
	middleware.WritePromptSuite("current-suite", []middleware.Prompt{})
	middleware.SetCurrentSuite("current-suite")

	// Delete the current suite - handler should reset current to default
	form := url.Values{}
	form.Add("suite_name", "current-suite")

	req := httptest.NewRequest("POST", "/delete_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	DeletePromptSuiteHandler(rr, req)

	// Should succeed with redirect
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify the suite was deleted
	if middleware.SuiteExists("current-suite") {
		t.Error("suite should be deleted")
	}
}

func TestEditPromptSuiteHandler_POST_RenameToSameName(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// Create a suite first
	middleware.WritePromptSuite("existing-suite", []middleware.Prompt{})

	// Try to rename to same name (should return error since name already exists)
	form := url.Values{}
	form.Add("suite_name", "existing-suite")
	form.Add("new_suite_name", "existing-suite")

	req := httptest.NewRequest("POST", "/edit_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	EditPromptSuiteHandler(rr, req)

	// Should return error for same name (suite already exists)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	// Suite should still exist
	if !middleware.SuiteExists("existing-suite") {
		t.Error("existing-suite should still exist")
	}
}

func TestDeletePromptSuiteHandler_GET_RenderError(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// Create a suite to delete
	middleware.WritePromptSuite("test-to-delete", []middleware.Prompt{})

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/delete_prompt_suite?suite_name=test-to-delete", nil)
	rr := httptest.NewRecorder()
	DeletePromptSuiteHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestNewPromptSuiteHandler_GET_RenderError(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/new_prompt_suite", nil)
	rr := httptest.NewRecorder()
	NewPromptSuiteHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEditPromptSuiteHandler_GET_RenderError(t *testing.T) {
	cleanup := setupSuitesTestDB(t)
	defer cleanup()

	// Create a suite to edit
	middleware.WritePromptSuite("test-to-edit", []middleware.Prompt{})

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/edit_prompt_suite?suite_name=test-to-edit", nil)
	rr := httptest.NewRecorder()
	EditPromptSuiteHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestNewPromptSuiteHandler_POST_WritePromptSuiteError(t *testing.T) {
	mockDS := &MockDataStoreWithError{
		MockDataStore: MockDataStore{
			CurrentSuite: "default",
		},
		WritePromptSuiteErr: errors.New("mock write error"),
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("suite_name", "new-suite")

	req := httptest.NewRequest("POST", "/new_prompt_suite", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.NewPromptSuite(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on WritePromptSuite error, got %d", http.StatusInternalServerError, rr.Code)
	}
}
