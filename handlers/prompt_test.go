package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"llm-tournament/middleware"

	_ "github.com/mattn/go-sqlite3"
)

// setupPromptTestDB creates a test database for prompt handler tests
func setupPromptTestDB(t *testing.T) func() {
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

func TestMovePromptHandler_POST_Success(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add multiple prompts
	for i := 0; i < 3; i++ {
		form := url.Values{}
		form.Add("prompt", "Prompt "+string(rune('A'+i)))
		req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		AddPromptHandler(httptest.NewRecorder(), req)
	}

	// Move prompt from index 0 to index 2
	moveForm := url.Values{}
	moveForm.Add("index", "0")
	moveForm.Add("new_index", "2")

	moveReq := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(moveForm.Encode()))
	moveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	moveRR := httptest.NewRecorder()
	MovePromptHandler(moveRR, moveReq)

	if moveRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, moveRR.Code)
	}
}

func TestMovePromptHandler_POST_InvalidIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	moveForm := url.Values{}
	moveForm.Add("index", "invalid")
	moveForm.Add("new_index", "1")

	moveReq := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(moveForm.Encode()))
	moveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	moveRR := httptest.NewRecorder()
	MovePromptHandler(moveRR, moveReq)

	if moveRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, moveRR.Code)
	}
}

func TestMovePromptHandler_POST_InvalidNewIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	moveForm := url.Values{}
	moveForm.Add("index", "0")
	moveForm.Add("new_index", "invalid")

	moveReq := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(moveForm.Encode()))
	moveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	moveRR := httptest.NewRecorder()
	MovePromptHandler(moveRR, moveReq)

	if moveRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, moveRR.Code)
	}
}

func TestMovePromptHandler_GET_InvalidIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/move_prompt?index=invalid", nil)
	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
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

func TestUpdatePromptsOrderHandler_ValidOrder(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// The UpdatePromptsOrder function expects database prompt IDs in the order array
	// which makes it complex to test without knowing the actual IDs generated.
	// For now, just test that the handler accepts valid JSON and redirects.
	// Invalid orders will log errors but still redirect.

	order := []int{}
	orderJSON, _ := json.Marshal(order)

	form := url.Values{}
	form.Add("order", string(orderJSON))

	orderReq := httptest.NewRequest("POST", "/update_prompts_order", strings.NewReader(form.Encode()))
	orderReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	orderRR := httptest.NewRecorder()
	UpdatePromptsOrderHandler(orderRR, orderReq)

	// Handler always redirects after processing
	if orderRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, orderRR.Code)
	}
}

func TestUpdatePromptsOrderHandler_EmptyOrder(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("order", "")

	orderReq := httptest.NewRequest("POST", "/update_prompts_order", strings.NewReader(form.Encode()))
	orderReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	orderRR := httptest.NewRecorder()
	UpdatePromptsOrderHandler(orderRR, orderReq)

	if orderRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, orderRR.Code)
	}
}

func TestUpdatePromptsOrderHandler_InvalidJSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("order", "not valid json")

	orderReq := httptest.NewRequest("POST", "/update_prompts_order", strings.NewReader(form.Encode()))
	orderReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	orderRR := httptest.NewRecorder()
	UpdatePromptsOrderHandler(orderRR, orderReq)

	if orderRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, orderRR.Code)
	}
}

func TestBulkDeletePromptsHandler_Success(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add some prompts
	for i := 0; i < 3; i++ {
		form := url.Values{}
		form.Add("prompt", "Bulk delete test "+string(rune('A'+i)))
		req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		AddPromptHandler(httptest.NewRecorder(), req)
	}

	// Bulk delete indices 0 and 2
	requestBody := map[string][]int{"indices": {0, 2}}
	jsonBody, _ := json.Marshal(requestBody)

	bulkReq := httptest.NewRequest("POST", "/bulk_delete_prompts", bytes.NewReader(jsonBody))
	bulkReq.Header.Set("Content-Type", "application/json")

	bulkRR := httptest.NewRecorder()
	BulkDeletePromptsHandler(bulkRR, bulkReq)

	if bulkRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, bulkRR.Code)
	}

	// Verify only 1 prompt remains
	prompts := middleware.ReadPrompts()
	if len(prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(prompts))
	}
}

func TestBulkDeletePromptsHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/bulk_delete_prompts", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestBulkDeletePromptsHandler_EmptyIndices(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add a prompt
	form := url.Values{}
	form.Add("prompt", "Test prompt")
	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddPromptHandler(httptest.NewRecorder(), req)

	// Bulk delete with empty indices
	requestBody := map[string][]int{"indices": {}}
	jsonBody, _ := json.Marshal(requestBody)

	bulkReq := httptest.NewRequest("POST", "/bulk_delete_prompts", bytes.NewReader(jsonBody))
	bulkReq.Header.Set("Content-Type", "application/json")

	bulkRR := httptest.NewRecorder()
	BulkDeletePromptsHandler(bulkRR, bulkReq)

	if bulkRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, bulkRR.Code)
	}
}

func TestBulkDeletePromptsPageHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/bulk_delete_prompts_page", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsPageHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestBulkDeletePromptsPageHandler_NoIndices(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/bulk_delete_prompts_page", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsPageHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestBulkDeletePromptsPageHandler_InvalidIndicesJSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/bulk_delete_prompts_page?indices=not_json", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsPageHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestPromptListHandler_InvalidOrderFilter(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/prompts?order_filter=invalid", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
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
	writer.Close()

	return body, writer.FormDataContentType()
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

func TestImportResultsHandler_POST_ValidJSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// First create prompts so results have targets
	prompts := []middleware.Prompt{
		{Text: "Prompt 1", Solution: "Solution 1"},
		{Text: "Prompt 2", Solution: "Solution 2"},
	}
	middleware.WritePrompts(prompts)

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
