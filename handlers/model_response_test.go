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
