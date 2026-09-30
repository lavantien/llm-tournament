package handlers

import (
	"errors"
	"llm-tournament/middleware"
	"llm-tournament/testutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// changeToProjectRootPrompts changes to project root for template tests
func changeToProjectRootPrompts(t *testing.T) func() {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}
	return func() {
		_ = os.Chdir(originalDir)
	}
}

// setupPromptTestDB creates a test database for prompt handler tests
func setupPromptTestDB(t *testing.T) func() {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	err := middleware.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	return func() {
		_ = middleware.CloseDB()
	}
}

func TestAddPromptHandler_Success(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("prompt", "Test prompt text")
	form.Add("solution", "Test solution")
	form.Add("profile", "")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddPromptHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify prompt was added
	prompts := middleware.ReadPrompts()
	if len(prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(prompts))
	}
	if prompts[0].Text != "Test prompt text" {
		t.Errorf("expected prompt text 'Test prompt text', got %q", prompts[0].Text)
	}
}

func TestAddPromptHandler_EmptyText(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("prompt", "")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddPromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAddPromptHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/add_prompt", nil)
	rr := httptest.NewRecorder()
	AddPromptHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestEditPromptHandler_POST_Success(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// First add a prompt
	addForm := url.Values{}
	addForm.Add("prompt", "Original prompt")

	addReq := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddPromptHandler(httptest.NewRecorder(), addReq)

	// Edit the prompt
	editForm := url.Values{}
	editForm.Add("index", "0")
	editForm.Add("prompt", "Edited prompt")
	editForm.Add("solution", "Edited solution")
	editForm.Add("profile", "")

	editReq := httptest.NewRequest("POST", "/edit_prompt", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditPromptHandler(editRR, editReq)

	if editRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, editRR.Code)
	}

	// Verify prompt was edited
	prompts := middleware.ReadPrompts()
	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(prompts))
	}
	if prompts[0].Text != "Edited prompt" {
		t.Errorf("expected prompt text 'Edited prompt', got %q", prompts[0].Text)
	}
}

func TestEditPromptHandler_POST_EmptyText(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// First add a prompt
	addForm := url.Values{}
	addForm.Add("prompt", "Original prompt")

	addReq := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddPromptHandler(httptest.NewRecorder(), addReq)

	// Try to edit with empty text
	editForm := url.Values{}
	editForm.Add("index", "0")
	editForm.Add("prompt", "")

	editReq := httptest.NewRequest("POST", "/edit_prompt", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditPromptHandler(editRR, editReq)

	if editRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, editRR.Code)
	}
}

func TestEditPromptHandler_POST_InvalidIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	editForm := url.Values{}
	editForm.Add("index", "invalid")
	editForm.Add("prompt", "New prompt")

	editReq := httptest.NewRequest("POST", "/edit_prompt", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditPromptHandler(editRR, editReq)

	if editRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, editRR.Code)
	}
}

func TestEditPromptHandler_GET_InvalidIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/edit_prompt?index=invalid", nil)
	rr := httptest.NewRecorder()
	EditPromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestEditPromptHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/edit_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	EditPromptHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestDeletePromptHandler_POST_Success(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// First add a prompt
	addForm := url.Values{}
	addForm.Add("prompt", "Prompt to delete")

	addReq := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddPromptHandler(httptest.NewRecorder(), addReq)

	// Delete the prompt
	deleteForm := url.Values{}
	deleteForm.Add("index", "0")

	deleteReq := httptest.NewRequest("POST", "/delete_prompt", strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	deleteRR := httptest.NewRecorder()
	DeletePromptHandler(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, deleteRR.Code)
	}

	// Verify prompt was deleted
	prompts := middleware.ReadPrompts()
	if len(prompts) != 0 {
		t.Errorf("expected 0 prompts, got %d", len(prompts))
	}
}

func TestDeletePromptHandler_POST_InvalidIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	deleteForm := url.Values{}
	deleteForm.Add("index", "not_a_number")

	deleteReq := httptest.NewRequest("POST", "/delete_prompt", strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	deleteRR := httptest.NewRecorder()
	DeletePromptHandler(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, deleteRR.Code)
	}
}

func TestDeletePromptHandler_GET_InvalidIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/delete_prompt?index=invalid", nil)
	rr := httptest.NewRecorder()
	DeletePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestDeletePromptHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/delete_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	DeletePromptHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestResetPromptsHandler_POST(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add some prompts
	for i := 0; i < 3; i++ {
		form := url.Values{}
		form.Add("prompt", "Prompt to reset "+string(rune('A'+i)))
		req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		AddPromptHandler(httptest.NewRecorder(), req)
	}

	// Verify we have 3 prompts
	prompts := middleware.ReadPrompts()
	if len(prompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(prompts))
	}

	// Reset prompts
	resetReq := httptest.NewRequest("POST", "/reset_prompts", nil)
	resetRR := httptest.NewRecorder()
	ResetPromptsHandler(resetRR, resetReq)

	if resetRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, resetRR.Code)
	}

	// Verify prompts were reset
	prompts = middleware.ReadPrompts()
	if len(prompts) != 0 {
		t.Errorf("expected 0 prompts after reset, got %d", len(prompts))
	}
}

func TestResetPromptsHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/reset_prompts", nil)
	rr := httptest.NewRecorder()
	ResetPromptsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestEditPrompt_GET_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodGet, "/edit_prompt?index=%zz", nil)
	rr := httptest.NewRecorder()
	handler.EditPrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestEditPrompt_POST_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/edit_prompt", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.EditPrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestDeletePrompt_GET_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodGet, "/delete_prompt?index=%zz", nil)
	rr := httptest.NewRecorder()
	handler.DeletePrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestDeletePrompt_POST_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/delete_prompt", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.DeletePrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestResetPromptsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/reset_prompts", nil)
	rr := httptest.NewRecorder()
	ResetPromptsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestDeletePromptHandler_GET_Success(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add a prompt first
	prompts := []middleware.Prompt{{Text: "Delete me"}}
	_ = middleware.WritePrompts(prompts)

	req := httptest.NewRequest("GET", "/delete_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	DeletePromptHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Delete me") {
		t.Error("expected prompt text in response body")
	}
}

func TestDeletePromptHandler_GET_OutOfRange(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add a prompt
	prompts := []middleware.Prompt{{Text: "Test"}}
	_ = middleware.WritePrompts(prompts)

	req := httptest.NewRequest("GET", "/delete_prompt?index=99", nil)
	rr := httptest.NewRecorder()
	DeletePromptHandler(rr, req)

	// Out of range index should not render template
	if rr.Code == http.StatusInternalServerError {
		t.Error("unexpected internal server error")
	}
}

func TestEditPromptHandler_GET_Success(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompt and profile
	_ = middleware.WriteProfiles([]middleware.Profile{{Name: "TestProfile"}})
	prompts := []middleware.Prompt{{Text: "Edit me", Profile: "TestProfile"}}
	_ = middleware.WritePrompts(prompts)

	req := httptest.NewRequest("GET", "/edit_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	EditPromptHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Edit me") {
		t.Error("expected prompt text in response body")
	}
}

func TestAddPromptHandler_WithProfile(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add a profile first
	_ = middleware.WriteProfiles([]middleware.Profile{{Name: "TestProfile"}})

	form := url.Values{}
	form.Add("prompt", "Prompt with profile")
	form.Add("profile", "TestProfile")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddPromptHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify prompt was added with profile
	prompts := middleware.ReadPrompts()
	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(prompts))
	}
	if prompts[0].Profile != "TestProfile" {
		t.Errorf("expected profile 'TestProfile', got %q", prompts[0].Profile)
	}
}

func TestResetPromptsHandler_GET_WithTemplate(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/reset_prompts", nil)
	rr := httptest.NewRecorder()
	ResetPromptsHandler(rr, req)

	// GET should render confirmation template
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestAddPromptHandler_WithSolution(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("prompt", "Test prompt")
	form.Add("solution", "Expected solution")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddPromptHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify prompt with solution was added
	prompts := middleware.ReadPrompts()
	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(prompts))
	}
	if prompts[0].Solution != "Expected solution" {
		t.Errorf("expected solution 'Expected solution', got %q", prompts[0].Solution)
	}
}

func TestAddPromptHandler_WithType(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("prompt", "Creative prompt")
	form.Add("type", "creative")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddPromptHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestResetPromptsHandler_GET_RenderError(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/reset_prompts", nil)
	rr := httptest.NewRecorder()
	ResetPromptsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestAddPromptHandler_WhitespaceOnlyText(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("prompt", "   \t\n   ")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddPromptHandler(rr, req)

	// Whitespace-only prompts may be accepted by the handler (redirects)
	// or rejected with bad request, depending on implementation
	if rr.Code != http.StatusSeeOther && rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d or %d for whitespace-only prompt, got %d",
			http.StatusSeeOther, http.StatusBadRequest, rr.Code)
	}
}

func TestAddPromptHandler_WritePromptsError(t *testing.T) {
	mockDS := &MockDataStoreWithError{
		MockDataStore: MockDataStore{
			Prompts:      []middleware.Prompt{},
			CurrentSuite: "test-suite",
		},
		WritePromptSuiteErr: errors.New("mock write error"),
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("prompt", "New test prompt")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.AddPrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeletePromptHandler_WritePromptsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
		},
		CurrentSuite: "test-suite",
		WritePromptsFunc: func(prompts []middleware.Prompt) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest("POST", "/delete_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	handler.DeletePrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEditPromptHandler_WritePromptsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Original prompt"},
		},
		CurrentSuite: "test-suite",
		WritePromptsFunc: func(prompts []middleware.Prompt) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("prompt", "Updated prompt text")

	req := httptest.NewRequest("POST", "/edit_prompt?index=0", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.EditPrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestResetPromptsHandler_WritePromptsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts:      []middleware.Prompt{{Text: "Test"}},
		CurrentSuite: "test-suite",
		WritePromptsFunc: func(prompts []middleware.Prompt) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest("POST", "/reset_prompts", nil)
	rr := httptest.NewRecorder()
	handler.ResetPrompts(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeletePromptHandler_GET_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
		},
		CurrentSuite: "test-suite",
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/delete_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	handler.DeletePrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEditPromptHandler_GET_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
		},
		CurrentSuite: "test-suite",
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/edit_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	handler.EditPrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestAddPromptHandler_ReadPromptSuiteError(t *testing.T) {
	mockDS := &MockDataStoreWithError{
		MockDataStore: MockDataStore{
			Prompts:      []middleware.Prompt{},
			CurrentSuite: "test-suite",
		},
		ReadPromptSuiteErr: errors.New("mock read error"),
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("prompt", "Test prompt")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.AddPrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on read error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestAddPrompt_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/add_prompt", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.AddPrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

type emptySuiteNameDataStore struct {
	MockDataStore
	ReadSuiteName  string
	WriteSuiteName string
}

func (m *emptySuiteNameDataStore) GetCurrentSuiteName() string { return "" }

func (m *emptySuiteNameDataStore) ReadPromptSuite(suiteName string) ([]middleware.Prompt, error) {
	m.ReadSuiteName = suiteName
	return m.MockDataStore.ReadPromptSuite(suiteName)
}

func (m *emptySuiteNameDataStore) WritePromptSuite(suiteName string, prompts []middleware.Prompt) error {
	m.WriteSuiteName = suiteName
	return m.MockDataStore.WritePromptSuite(suiteName, prompts)
}

func TestAddPromptHandler_EmptySuiteName(t *testing.T) {
	mockDS := &emptySuiteNameDataStore{
		MockDataStore: MockDataStore{
			Prompts: []middleware.Prompt{},
		},
	}
	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("prompt", "Test prompt")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.AddPrompt(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d (redirect), got %d", http.StatusSeeOther, rr.Code)
	}
	if mockDS.ReadSuiteName != "default" {
		t.Errorf("expected ReadPromptSuite to use default suite, got %q", mockDS.ReadSuiteName)
	}
	if mockDS.WriteSuiteName != "default" {
		t.Errorf("expected WritePromptSuite to use default suite, got %q", mockDS.WriteSuiteName)
	}
	if len(mockDS.Prompts) != 1 {
		t.Errorf("expected 1 prompt written, got %d", len(mockDS.Prompts))
	}
}
