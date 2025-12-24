package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"llm-tournament/middleware"

	_ "github.com/mattn/go-sqlite3"
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
