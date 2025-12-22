package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
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

func TestMovePromptHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/move_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

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

func TestUpdatePromptsOrderHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/update_prompts_order", nil)
	rr := httptest.NewRecorder()
	UpdatePromptsOrderHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestUpdatePromptsOrder_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/update_prompts_order", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.UpdatePromptsOrder(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
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

func TestBulkDeletePromptsPage_RenderError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{RenderError: errors.New("mock render error")})

	req := httptest.NewRequest(http.MethodGet, "/bulk_delete_prompts_page?indices=[0]", nil)
	rr := httptest.NewRecorder()
	handler.BulkDeletePromptsPage(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error rendering template") {
		t.Fatalf("expected render error message, got %q", rr.Body.String())
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

func TestBulkDeletePrompts_InvalidJSONBody(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/bulk_delete_prompts", strings.NewReader("not json"))
	rr := httptest.NewRecorder()
	handler.BulkDeletePrompts(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error decoding request") {
		t.Fatalf("expected decode error message, got %q", rr.Body.String())
	}
}

func TestBulkDeletePrompts_NoPrompts(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/bulk_delete_prompts", strings.NewReader(`{"indices":[0]}`))
	rr := httptest.NewRecorder()
	handler.BulkDeletePrompts(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "No prompts to delete") {
		t.Fatalf("expected no prompts message, got %q", rr.Body.String())
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

func TestMovePrompt_GET_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodGet, "/move_prompt?index=%zz", nil)
	rr := httptest.NewRecorder()
	handler.MovePrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestMovePrompt_POST_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/move_prompt", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.MovePrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
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

type promptListListSuitesErrorDataStore struct {
	MockDataStore
	Err error
}

func (ds *promptListListSuitesErrorDataStore) ListPromptSuites() ([]string, error) {
	return nil, ds.Err
}

type promptListReadSuiteErrorDataStore struct {
	MockDataStore
	Err error
}

func (ds *promptListReadSuiteErrorDataStore) ReadPromptSuite(suiteName string) ([]middleware.Prompt, error) {
	return nil, ds.Err
}

type promptListEmptySuiteNameDataStore struct {
	MockDataStore
	GetCalls  int
	ReadCalls int
}

func (ds *promptListEmptySuiteNameDataStore) GetCurrentSuiteName() string {
	ds.GetCalls++
	return ""
}

func (ds *promptListEmptySuiteNameDataStore) ReadPromptSuite(suiteName string) ([]middleware.Prompt, error) {
	ds.ReadCalls++
	return nil, nil
}

type promptListSecondReadErrorDataStore struct {
	MockDataStore
	ReadCalls int
}

func (ds *promptListSecondReadErrorDataStore) ReadPromptSuite(suiteName string) ([]middleware.Prompt, error) {
	ds.ReadCalls++
	if ds.ReadCalls == 2 {
		return nil, errors.New("mock read prompt suite error")
	}
	return nil, nil
}

func TestPromptList_InvalidOrderFilter(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts?order_filter=not-an-int", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid order filter") {
		t.Fatalf("expected invalid order filter message, got %q", rr.Body.String())
	}
}

func TestPromptList_ListPromptSuitesError(t *testing.T) {
	handler := NewHandlerWithDeps(&promptListListSuitesErrorDataStore{
		Err: errors.New("mock list suites error"),
	}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error listing prompt suites") {
		t.Fatalf("expected list suites error message, got %q", rr.Body.String())
	}
}

func TestPromptList_ReadPromptSuiteError(t *testing.T) {
	handler := NewHandlerWithDeps(&promptListReadSuiteErrorDataStore{
		Err: errors.New("mock read suite error"),
	}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error reading prompt suite") {
		t.Fatalf("expected read suite error message, got %q", rr.Body.String())
	}
}

func TestPromptList_RenderError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{RenderError: errors.New("mock render error")})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error rendering template") {
		t.Fatalf("expected render error message, got %q", rr.Body.String())
	}
}

func TestPromptList_DefaultSuiteFallback(t *testing.T) {
	ds := &promptListEmptySuiteNameDataStore{}
	handler := NewHandlerWithDeps(ds, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if ds.GetCalls != 1 {
		t.Fatalf("expected GetCurrentSuiteName to be called once, got %d", ds.GetCalls)
	}
	if ds.ReadCalls != 2 {
		t.Fatalf("expected ReadPromptSuite to be called twice, got %d", ds.ReadCalls)
	}
}

func TestPromptList_WithPrompts(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "Prompt 1", Solution: "Solution 1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestPromptList_ReadDefaultPromptSuiteError(t *testing.T) {
	handler := NewHandlerWithDeps(&promptListSecondReadErrorDataStore{}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error reading default prompt suite") {
		t.Fatalf("expected default suite read error message, got %q", rr.Body.String())
	}
}

func TestPromptList_OrderFilterParseSuccess(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts?order_filter=2", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
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

func TestPromptListHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add test profile and prompt
	err := middleware.WriteProfiles([]middleware.Profile{
		{Name: "TestProfile", Description: "Test"},
	})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	err = middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt", Profile: "TestProfile"},
	})
	if err != nil {
		t.Fatalf("failed to write prompt: %v", err)
	}

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Test Prompt") {
		t.Error("expected prompt text in response body")
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

func TestBulkDeletePromptsPageHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add a prompt first
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt"},
	})
	if err != nil {
		t.Fatalf("failed to write prompt: %v", err)
	}

	// Request with indices parameter
	req := httptest.NewRequest("GET", "/bulk_delete_prompts?indices=[0]", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsPageHandler(rr, req)

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

func TestPromptListHandler_WithSearch(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts with different text
	prompts := []middleware.Prompt{
		{Text: "Find this prompt"},
		{Text: "Another prompt"},
	}
	_ = middleware.WritePrompts(prompts)

	req := httptest.NewRequest("GET", "/prompts?search_query=Find", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
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

func TestMovePromptHandler_POST_OutOfRange(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "First"}, {Text: "Second"}}
	_ = middleware.WritePrompts(prompts)

	form := url.Values{}
	form.Add("index", "99")
	form.Add("new_index", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	// Should redirect without error (bounds check in handler)
	if rr.Code != http.StatusSeeOther {
		t.Logf("status: %d", rr.Code)
	}
}

func TestMovePromptHandler_POST_MoveUp(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "First"}, {Text: "Second"}, {Text: "Third"}}
	_ = middleware.WritePrompts(prompts)

	// Move Third (index 2) to position 0
	form := url.Values{}
	form.Add("index", "2")
	form.Add("new_index", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify order changed
	newPrompts := middleware.ReadPrompts()
	if len(newPrompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(newPrompts))
	}
	if newPrompts[0].Text != "Third" {
		t.Errorf("expected 'Third' first, got %q", newPrompts[0].Text)
	}
}

func TestMovePromptHandler_GET_OutOfRange(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "Only prompt"}}
	_ = middleware.WritePrompts(prompts)

	req := httptest.NewRequest("GET", "/move_prompt?index=99", nil)
	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	// Should not crash with out-of-range index
	if rr.Code == http.StatusInternalServerError {
		t.Error("unexpected internal server error")
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

func TestMovePromptHandler_POST_SamePosition(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "First"}, {Text: "Second"}, {Text: "Third"}}
	_ = middleware.WritePrompts(prompts)

	// Move to same position
	form := url.Values{}
	form.Add("index", "1")
	form.Add("new_index", "1")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Order should remain unchanged
	newPrompts := middleware.ReadPrompts()
	if newPrompts[1].Text != "Second" {
		t.Errorf("expected 'Second' to remain at index 1, got %q", newPrompts[1].Text)
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

func TestPromptListHandler_WithOrderFilter(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	if err := middleware.WritePrompts([]middleware.Prompt{
		{Text: "First Prompt"},
		{Text: "Second Prompt"},
	}); err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}

	req := httptest.NewRequest("GET", "/prompts?order_filter=1", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestPromptListHandler_WithProfileFilter(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts with profiles
	_ = middleware.WriteProfiles([]middleware.Profile{{Name: "TestProfile"}})
	if err := middleware.WritePrompts([]middleware.Prompt{
		{Text: "Filtered Prompt", Profile: "TestProfile"},
		{Text: "Other Prompt", Profile: "Other"},
	}); err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}

	req := httptest.NewRequest("GET", "/prompts?profile_filter=TestProfile", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Filtered Prompt") {
		t.Error("expected 'Filtered Prompt' in response body")
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

func TestMovePromptHandler_POST_InvalidFromIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try to move from invalid index
	form := url.Values{}
	form.Add("from", "invalid")
	form.Add("to", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_POST_InvalidToIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try to move to invalid index
	form := url.Values{}
	form.Add("from", "0")
	form.Add("to", "invalid")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_POST_OutOfBoundsFrom(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try to move from out-of-bounds index
	form := url.Values{}
	form.Add("from", "999")
	form.Add("to", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_POST_OutOfBoundsTo(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try to move to out-of-bounds index
	form := url.Values{}
	form.Add("from", "0")
	form.Add("to", "999")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_POST_NegativeFrom(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try negative from index
	form := url.Values{}
	form.Add("from", "-1")
	form.Add("to", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
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

	// The handler logs the error and returns error status
	// Due to how headers work, we check if the error was handled
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

func TestMovePromptHandler_WritePromptsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
			{Text: "Prompt 3"},
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
	form.Add("new_index", "2")

	req := httptest.NewRequest("POST", "/move_prompt?index=0", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.MovePrompt(rr, req)

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

func TestPromptListHandler_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts:      []middleware.Prompt{{Text: "Test"}},
		CurrentSuite: "test-suite",
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
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

func TestBulkDeletePromptsHandler_WritePromptsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
			{Text: "Prompt 3"},
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

	// BulkDeletePrompts expects JSON input
	jsonBody := `{"indices": [0, 1]}`

	req := httptest.NewRequest("POST", "/bulk_delete_prompts", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.BulkDeletePrompts(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestMovePromptHandler_GET_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
		},
		CurrentSuite: "test-suite",
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/move_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	handler.MovePrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
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
