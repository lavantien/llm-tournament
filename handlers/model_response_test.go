package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"llm-tournament/middleware"
)

// setupModelResponseTestDB creates a test database for model response handler tests
func setupModelResponseTestDB(t *testing.T) func() {
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

// TestSaveModelResponseHandler_MethodNotAllowed verifies GET requests are rejected
func TestSaveModelResponseHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupModelResponseTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/save_model_response", nil)
	rr := httptest.NewRecorder()
	SaveModelResponseHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// TestSaveModelResponseHandler_MissingModelID verifies requests without model_id are rejected
func TestSaveModelResponseHandler_MissingModelID(t *testing.T) {
	cleanup := setupModelResponseTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/save_model_response", nil)
	rr := httptest.NewRecorder()
	SaveModelResponseHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestSaveModelResponseHandler_MissingPromptID verifies requests without prompt_id are rejected
func TestSaveModelResponseHandler_MissingPromptID(t *testing.T) {
	cleanup := setupModelResponseTestDB(t)
	defer cleanup()

	reqBody := map[string]int{
		"model_id": 1,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/save_model_response", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	SaveModelResponseHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestSaveModelResponseHandler_MissingResponseText verifies requests without response_text are rejected
func TestSaveModelResponseHandler_MissingResponseText(t *testing.T) {
	cleanup := setupModelResponseTestDB(t)
	defer cleanup()

	reqBody := map[string]int{
		"model_id":  1,
		"prompt_id": 1,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/save_model_response", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	SaveModelResponseHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if rr.Body.String() != "response_text is required\n" {
		t.Errorf("expected body 'response_text is required\n', got '%s'", rr.Body.String())
	}
}

// TestSaveModelResponseHandler_ValidInput saves response successfully
func TestSaveModelResponseHandler_ValidInput(t *testing.T) {
	cleanup := setupModelResponseTestDB(t)
	defer cleanup()

	// First, create a test model and prompt
	db := middleware.GetDB()
	var modelID, promptID int
	db.QueryRow("INSERT INTO models (name, suite_id) VALUES ('test-model', 1) RETURNING id").Scan(&modelID)
	db.QueryRow("INSERT INTO prompts (text, suite_id, display_order, type) VALUES ('test prompt', 1, 1, 'objective') RETURNING id").Scan(&promptID)

	reqBody := map[string]interface{}{
		"model_id":      modelID,
		"prompt_id":     promptID,
		"response_text": "This is a test response",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/save_model_response", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	SaveModelResponseHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// Verify the response was saved
	var responseText string
	err := db.QueryRow("SELECT response_text FROM model_responses WHERE model_id = ? AND prompt_id = ?", modelID, promptID).Scan(&responseText)
	if err != nil {
		t.Errorf("failed to query saved response: %v", err)
	}
	if responseText != "This is a test response" {
		t.Errorf("expected response text 'This is a test response', got '%s'", responseText)
	}
}

// TestSaveModelResponseHandler_UpdateExisting updates existing response
func TestSaveModelResponseHandler_UpdateExisting(t *testing.T) {
	cleanup := setupModelResponseTestDB(t)
	defer cleanup()

	// First, create a test model and prompt
	db := middleware.GetDB()
	var modelID, promptID int
	db.QueryRow("INSERT INTO models (name, suite_id) VALUES ('test-model', 1) RETURNING id").Scan(&modelID)
	db.QueryRow("INSERT INTO prompts (text, suite_id, display_order, type) VALUES ('test prompt', 1, 1, 'objective') RETURNING id").Scan(&promptID)

	// Insert initial response
	_, _ = db.Exec("INSERT INTO model_responses (model_id, prompt_id, response_text, response_source) VALUES (?, ?, 'original', 'manual')", modelID, promptID)

	// Update with new response
	reqBody := map[string]interface{}{
		"model_id":      modelID,
		"prompt_id":     promptID,
		"response_text": "updated response",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/save_model_response", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	SaveModelResponseHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify the response was updated
	var responseText string
	err := db.QueryRow("SELECT response_text FROM model_responses WHERE model_id = ? AND prompt_id = ?", modelID, promptID).Scan(&responseText)
	if err != nil {
		t.Errorf("failed to query saved response: %v", err)
	}
	if responseText != "updated response" {
		t.Errorf("expected response text 'updated response', got '%s'", responseText)
	}
}
