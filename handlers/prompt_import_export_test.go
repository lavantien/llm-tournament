package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"llm-tournament/middleware"
	"llm-tournament/testutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestExportPromptsHandler(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add some prompts
	form := url.Values{}
	form.Add("prompt", "Export test prompt")
	form.Add("solution", "Export test solution")

	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddPromptHandler(httptest.NewRecorder(), req)

	// Export prompts
	exportReq := httptest.NewRequest("GET", "/export_prompts", nil)
	exportRR := httptest.NewRecorder()
	ExportPromptsHandler(exportRR, exportReq)

	if exportRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, exportRR.Code)
	}

	// Verify JSON content
	var prompts []middleware.Prompt
	err := json.Unmarshal(exportRR.Body.Bytes(), &prompts)
	if err != nil {
		t.Fatalf("failed to unmarshal exported JSON: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("expected 1 prompt in export, got %d", len(prompts))
	}
}

func TestExportPromptsHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/export_prompts", nil)
	rr := httptest.NewRecorder()
	ExportPromptsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// Helper to create multipart form with file
func createMultipartFormFile(t *testing.T, fieldname, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldname, filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}
	_ = writer.Close()

	return body, writer.FormDataContentType()
}

func TestImportPromptsHandler_MethodNotAllowed(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{},
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest(http.MethodPut, "/import_prompts", nil)
	rr := httptest.NewRecorder()
	handler.ImportPrompts(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestImportPromptsHandler_POST_NoFile(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// POST without file should redirect to error
	req := httptest.NewRequest("POST", "/import_prompts", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	rr := httptest.NewRecorder()
	ImportPromptsHandler(rr, req)

	// Should redirect to import_error
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for no file, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestImportPromptsHandler_POST_ValidJSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	prompts := []middleware.Prompt{
		{Text: "Imported Prompt 1", Solution: "Solution 1"},
		{Text: "Imported Prompt 2", Solution: "Solution 2"},
	}
	jsonData, _ := json.Marshal(prompts)

	body, contentType := createMultipartFormFile(t, "prompts_file", "prompts.json", jsonData)

	req := httptest.NewRequest("POST", "/import_prompts", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	ImportPromptsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify prompts were imported
	importedPrompts := middleware.ReadPrompts()
	if len(importedPrompts) != 2 {
		t.Errorf("expected 2 imported prompts, got %d", len(importedPrompts))
	}
}

func TestImportPromptsHandler_POST_InvalidJSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	invalidJSON := []byte("not valid json")
	body, contentType := createMultipartFormFile(t, "prompts_file", "prompts.json", invalidJSON)

	req := httptest.NewRequest("POST", "/import_prompts", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	ImportPromptsHandler(rr, req)

	// Should redirect to import_error
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for invalid JSON, got %d", http.StatusSeeOther, rr.Code)
	}

	location := rr.Header().Get("Location")
	if !strings.Contains(location, "import_error") {
		t.Errorf("expected redirect to import_error, got %q", location)
	}
}

func TestImportPromptsHandler_ReadAllError(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	handler := &Handler{
		DataStore: &MockDataStore{},
		Renderer:  &MockRenderer{},
	}

	prompts := []middleware.Prompt{{Text: "Imported Prompt"}}
	jsonData, _ := json.Marshal(prompts)

	body, contentType := createMultipartFormFile(t, "prompts_file", "prompts.json", jsonData)
	req := httptest.NewRequest("POST", "/import_prompts", body)
	req.Header.Set("Content-Type", contentType)

	original := readAll
	readAll = func(io.Reader) ([]byte, error) { return nil, errors.New("mock readall error") }
	t.Cleanup(func() { readAll = original })

	rr := httptest.NewRecorder()
	handler.ImportPrompts(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error reading file") {
		t.Fatalf("expected error message, got %q", rr.Body.String())
	}
}

func TestImportResultsHandler_MethodNotAllowed(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{},
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest(http.MethodPut, "/import_results", nil)
	rr := httptest.NewRecorder()
	handler.ImportResults(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestImportResultsHandler_POST_NoFile(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/import_results", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	rr := httptest.NewRecorder()
	ImportResultsHandler(rr, req)

	// Should redirect to import_error
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for no file, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestImportResultsHandler_ReadAllError(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	handler := &Handler{
		DataStore: &MockDataStore{},
		Renderer:  &MockRenderer{},
	}

	results := map[string]middleware.Result{"Model1": {Scores: []int{80}}}
	jsonData, _ := json.Marshal(results)

	body, contentType := createMultipartFormFile(t, "results_file", "results.json", jsonData)
	req := httptest.NewRequest("POST", "/import_results", body)
	req.Header.Set("Content-Type", contentType)

	original := readAll
	readAll = func(io.Reader) ([]byte, error) { return nil, errors.New("mock readall error") }
	t.Cleanup(func() { readAll = original })

	rr := httptest.NewRecorder()
	handler.ImportResults(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error reading file") {
		t.Fatalf("expected error message, got %q", rr.Body.String())
	}
}

func TestImportResultsHandler_WriteResultsError(t *testing.T) {
	expectedErr := errors.New("mock write results error")
	handler := &Handler{
		DataStore: &MockDataStore{
			Prompts:      []middleware.Prompt{{Text: "Prompt 1"}},
			CurrentSuite: "test-suite",
			WriteResultsFunc: func(suiteName string, results map[string]middleware.Result) error {
				if suiteName != "test-suite" {
					t.Fatalf("expected suite name %q, got %q", "test-suite", suiteName)
				}
				return expectedErr
			},
		},
		Renderer: &MockRenderer{},
	}

	results := map[string]middleware.Result{"Model1": {Scores: []int{80}}}
	jsonData, _ := json.Marshal(results)

	body, contentType := createMultipartFormFile(t, "results_file", "results.json", jsonData)
	req := httptest.NewRequest("POST", "/import_results", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	handler.ImportResults(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error writing results") {
		t.Fatalf("expected error message, got %q", rr.Body.String())
	}
}

func TestImportResultsHandler_POST_ValidJSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// First create prompts so results have targets
	prompts := []middleware.Prompt{
		{Text: "Prompt 1", Solution: "Solution 1"},
		{Text: "Prompt 2", Solution: "Solution 2"},
	}
	_ = middleware.WritePrompts(prompts)

	results := map[string]middleware.Result{
		"Model A": {Scores: []int{80, 60}},
		"Model B": {Scores: []int{100, 40}},
	}
	jsonData, _ := json.Marshal(results)

	body, contentType := createMultipartFormFile(t, "results_file", "results.json", jsonData)

	req := httptest.NewRequest("POST", "/import_results", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	ImportResultsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestImportResultsHandler_POST_InvalidJSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	invalidJSON := []byte("not valid json")
	body, contentType := createMultipartFormFile(t, "results_file", "results.json", invalidJSON)

	req := httptest.NewRequest("POST", "/import_results", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	ImportResultsHandler(rr, req)

	// Should redirect to import_error
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for invalid JSON, got %d", http.StatusSeeOther, rr.Code)
	}

	location := rr.Header().Get("Location")
	if !strings.Contains(location, "import_error") {
		t.Errorf("expected redirect to import_error, got %q", location)
	}
}

func TestExportPromptsHandler_GET_JSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "Export me"}}
	_ = middleware.WritePrompts(prompts)

	req := httptest.NewRequest("GET", "/export_prompts", nil)
	rr := httptest.NewRecorder()
	ExportPromptsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Export me") {
		t.Error("expected prompt text in response body")
	}
}

func TestImportResultsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/import_results", nil)
	rr := httptest.NewRecorder()
	ImportResultsHandler(rr, req)

	// Should render the import form template
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestImportResultsHandler_GET_WithReturnTo(t *testing.T) {
	mockRenderer := &dataCaptureRenderer{}

	handler := &Handler{
		DataStore: &MockDataStore{},
		Renderer:  mockRenderer,
	}

	// Test with return_to query parameter
	req := httptest.NewRequest("GET", "/import_results?return_to=/profiles", nil)
	rr := httptest.NewRecorder()
	handler.ImportResults(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	dataMap, ok := mockRenderer.CapturedData.(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string, got %T", mockRenderer.CapturedData)
	}

	if dataMap["ReturnURL"] != "/profiles" {
		t.Errorf("expected ReturnURL \"/profiles\", got %q", dataMap["ReturnURL"])
	}
}

func TestImportPromptsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/import_prompts", nil)
	rr := httptest.NewRecorder()
	ImportPromptsHandler(rr, req)

	// Should render the import form template
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestImportResultsHandler_EmptyResults(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Create empty results JSON
	emptyResults := map[string]middleware.Result{}
	jsonData, _ := json.Marshal(emptyResults)

	body, contentType := createMultipartFormFile(t, "results_file", "results.json", jsonData)

	req := httptest.NewRequest("POST", "/import_results", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	ImportResultsHandler(rr, req)

	// Should redirect to import_error due to empty results
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	location := rr.Header().Get("Location")
	if !strings.Contains(location, "import_error") {
		t.Errorf("expected redirect to import_error, got %q", location)
	}
}

func TestImportResultsHandler_ScoresExtended(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Create more prompts than the imported scores
	prompts := []middleware.Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
		{Text: "Prompt 3"},
	}
	_ = middleware.WritePrompts(prompts)

	// Create results with fewer scores than prompts
	results := map[string]middleware.Result{
		"Model1": {Scores: []int{80}}, // Only 1 score but 3 prompts
	}
	jsonData, _ := json.Marshal(results)

	body, contentType := createMultipartFormFile(t, "results_file", "results.json", jsonData)

	req := httptest.NewRequest("POST", "/import_results", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	ImportResultsHandler(rr, req)

	// Should succeed
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify results were extended
	importedResults := middleware.ReadResults()
	if result, exists := importedResults["Model1"]; exists {
		if len(result.Scores) != 3 {
			t.Errorf("expected 3 scores after extension, got %d", len(result.Scores))
		}
	} else {
		t.Error("expected Model1 to exist in results")
	}
}

func TestExportPromptsHandler_Empty(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// No prompts
	req := httptest.NewRequest("GET", "/export_prompts", nil)
	rr := httptest.NewRecorder()
	ExportPromptsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Should return null or empty JSON array for no prompts
	body := strings.TrimSpace(rr.Body.String())
	if body != "null" && body != "[]" {
		t.Errorf("expected null or empty JSON array, got %q", body)
	}
}

func TestImportPromptsHandler_POST_EmptyArray(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Import an empty array - should redirect to import_error
	emptyJSON := []byte("[]")
	body, contentType := createMultipartFormFile(t, "prompts_file", "prompts.json", emptyJSON)

	req := httptest.NewRequest("POST", "/import_prompts", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	ImportPromptsHandler(rr, req)

	// Should redirect to import_error because no prompts found
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for empty array, got %d", http.StatusSeeOther, rr.Code)
	}

	location := rr.Header().Get("Location")
	if !strings.Contains(location, "import_error") {
		t.Errorf("expected redirect to import_error for empty array, got %q", location)
	}
}

func TestImportResultsHandler_POST_EmptyResults(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Import an empty results object - should redirect to import_error
	emptyJSON := []byte("{}")
	body, contentType := createMultipartFormFile(t, "results_file", "results.json", emptyJSON)

	req := httptest.NewRequest("POST", "/import_results", body)
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	ImportResultsHandler(rr, req)

	// Should redirect to import_error because no results found
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for empty results, got %d", http.StatusSeeOther, rr.Code)
	}

	location := rr.Header().Get("Location")
	if !strings.Contains(location, "import_error") {
		t.Errorf("expected redirect to import_error for empty results, got %q", location)
	}
}

func TestExportPromptsHandler_WithPrompts(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add some prompts with all fields
	if err := middleware.WritePrompts([]middleware.Prompt{
		{Text: "Prompt 1", Solution: "Solution 1", Profile: "Profile1"},
		{Text: "Prompt 2", Solution: "Solution 2", Profile: "Profile2"},
	}); err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}

	req := httptest.NewRequest("GET", "/export_prompts", nil)
	rr := httptest.NewRecorder()
	ExportPromptsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}

	// Verify JSON is valid and contains prompts
	var prompts []middleware.Prompt
	err := json.Unmarshal(rr.Body.Bytes(), &prompts)
	if err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
	if len(prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(prompts))
	}

	// Verify Content-Disposition header for download
	disposition := rr.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "attachment") {
		t.Errorf("expected attachment disposition, got %q", disposition)
	}
}

func TestImportPromptsHandler_GET_RenderError(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/import_prompts", nil)
	rr := httptest.NewRecorder()
	ImportPromptsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestImportResultsHandler_GET_RenderError(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/import_results", nil)
	rr := httptest.NewRecorder()
	ImportResultsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// Additional tests for low-coverage functions

func TestExportPromptsHandler_WriteError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Test prompt 1"},
			{Text: "Test prompt 2"},
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	// Use FailingResponseWriter to simulate write error
	rr := httptest.NewRecorder()
	failingWriter := &FailingResponseWriter{
		ResponseWriter: rr,
		WriteError:     errors.New("mock write error"),
	}

	req := httptest.NewRequest("GET", "/export_prompts", nil)
	handler.ExportPrompts(failingWriter, req)

	// The failing writer must have been invoked and the handler must have
	// reported the failure instead of a successful JSON download
	if failingWriter.WritesAttempted == 0 {
		t.Error("expected the handler to attempt writing the export payload")
	}
	if !failingWriter.HeaderWritten {
		t.Error("expected the error response path to write a header")
	}
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Errorf("expected no export payload on write error, got %q", rr.Body.String())
	}
}

func TestImportPromptsHandler_WritePromptsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts:      []middleware.Prompt{},
		CurrentSuite: "test-suite",
		WritePromptsFunc: func(prompts []middleware.Prompt) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	// Create multipart form with JSON file
	jsonContent := `[{"text": "Imported prompt", "solution": "solution"}]`
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("prompts_file", "prompts.json")
	_, _ = part.Write([]byte(jsonContent))
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/import_prompts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler.ImportPrompts(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestImportResultsHandler_POST_ShorterScores(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
			{Text: "Prompt 3"},
		},
		CurrentSuite: "test-suite",
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	// Create results with shorter scores than prompts
	results := map[string]middleware.Result{
		"Model1": {Scores: []int{80}}, // Only 1 score for 3 prompts
	}
	jsonData, _ := json.Marshal(results)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("results_file", "results.json")
	_, _ = part.Write(jsonData)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/import_results", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler.ImportResults(rr, req)

	// Should succeed with redirect (scores should be padded)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d (redirect), got %d", http.StatusSeeOther, rr.Code)
	}
}
