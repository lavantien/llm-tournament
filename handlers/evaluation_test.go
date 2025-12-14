package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"llm-tournament/middleware"

	_ "github.com/mattn/go-sqlite3"
)

// setupEvaluationTestDB creates a test database for evaluation handler tests
func setupEvaluationTestDB(t *testing.T) func() {
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

func TestEvaluateAllHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/evaluate/all", nil)
	rr := httptest.NewRecorder()
	EvaluateAllHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestEvaluateModelHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/evaluate/model?id=1", nil)
	rr := httptest.NewRecorder()
	EvaluateModelHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestEvaluateModelHandler_MissingID(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/evaluate/model", nil)
	rr := httptest.NewRecorder()
	EvaluateModelHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestEvaluateModelHandler_InvalidID(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/evaluate/model?id=invalid", nil)
	rr := httptest.NewRecorder()
	EvaluateModelHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestEvaluatePromptHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/evaluate/prompt?id=1", nil)
	rr := httptest.NewRecorder()
	EvaluatePromptHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestEvaluatePromptHandler_MissingID(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/evaluate/prompt", nil)
	rr := httptest.NewRecorder()
	EvaluatePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestEvaluatePromptHandler_InvalidID(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/evaluate/prompt?id=invalid", nil)
	rr := httptest.NewRecorder()
	EvaluatePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestEvaluationProgressHandler_MissingID(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/evaluation/progress", nil)
	rr := httptest.NewRecorder()
	EvaluationProgressHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestEvaluationProgressHandler_InvalidID(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/evaluation/progress?id=invalid", nil)
	rr := httptest.NewRecorder()
	EvaluationProgressHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCancelEvaluationHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/evaluation/cancel?id=1", nil)
	rr := httptest.NewRecorder()
	CancelEvaluationHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestCancelEvaluationHandler_MissingID(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/evaluation/cancel", nil)
	rr := httptest.NewRecorder()
	CancelEvaluationHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCancelEvaluationHandler_InvalidID(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/evaluation/cancel?id=invalid", nil)
	rr := httptest.NewRecorder()
	CancelEvaluationHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
