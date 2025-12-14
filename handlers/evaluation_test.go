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

func TestInitEvaluator(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Store original globalEvaluator
	originalEvaluator := globalEvaluator
	defer func() { globalEvaluator = originalEvaluator }()

	// Test initialization
	db := middleware.GetDB()
	InitEvaluator(db)

	if globalEvaluator == nil {
		t.Error("expected globalEvaluator to be initialized")
	}
}

func TestInitEvaluator_WithCustomURL(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Store original globalEvaluator
	originalEvaluator := globalEvaluator
	defer func() { globalEvaluator = originalEvaluator }()

	// Set custom python service URL
	middleware.SetSetting("python_service_url", "http://custom:8001")

	// Test initialization
	db := middleware.GetDB()
	InitEvaluator(db)

	if globalEvaluator == nil {
		t.Error("expected globalEvaluator to be initialized")
	}
}

func TestEvaluateAllHandler_WithEvaluator(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	req := httptest.NewRequest("POST", "/evaluate/all", nil)
	rr := httptest.NewRecorder()
	EvaluateAllHandler(rr, req)

	// Should succeed (creates a job with 0 prompts/models)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestEvaluateModelHandler_WithEvaluator(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	// Add a model directly using SQL
	suiteID, _ := middleware.GetSuiteID(middleware.GetCurrentSuiteName())
	_, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "TestModel", suiteID)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	req := httptest.NewRequest("POST", "/evaluate/model?id=1", nil)
	rr := httptest.NewRecorder()
	EvaluateModelHandler(rr, req)

	// Should succeed
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestEvaluatePromptHandler_WithEvaluator(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	// Add a prompt first
	middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	req := httptest.NewRequest("POST", "/evaluate/prompt?id=1", nil)
	rr := httptest.NewRecorder()
	EvaluatePromptHandler(rr, req)

	// Should succeed
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestEvaluationProgressHandler_WithValidJob(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	// Create a job directly in the database with all required fields
	suiteID, _ := middleware.GetSuiteID(middleware.GetCurrentSuiteName())
	_, err := db.Exec(`
		INSERT INTO evaluation_jobs (suite_id, job_type, target_id, status, progress_current, progress_total, error_message)
		VALUES (?, 'all', 0, 'running', 5, 10, '')
	`, suiteID)
	if err != nil {
		t.Fatalf("failed to insert job: %v", err)
	}

	req := httptest.NewRequest("GET", "/evaluation/progress?id=1", nil)
	rr := httptest.NewRecorder()
	EvaluationProgressHandler(rr, req)

	// Should return OK with job info
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestEvaluationProgressHandler_JobNotFound(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	req := httptest.NewRequest("GET", "/evaluation/progress?id=999", nil)
	rr := httptest.NewRecorder()
	EvaluationProgressHandler(rr, req)

	// Should return error for non-existent job
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestCancelEvaluationHandler_WithValidJob(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	// Create a job directly in the database with all required fields
	suiteID, _ := middleware.GetSuiteID(middleware.GetCurrentSuiteName())
	_, err := db.Exec(`
		INSERT INTO evaluation_jobs (suite_id, job_type, target_id, status, progress_current, progress_total, error_message)
		VALUES (?, 'all', 0, 'running', 5, 10, '')
	`, suiteID)
	if err != nil {
		t.Fatalf("failed to insert job: %v", err)
	}

	req := httptest.NewRequest("POST", "/evaluation/cancel?id=1", nil)
	rr := httptest.NewRecorder()
	CancelEvaluationHandler(rr, req)

	// Cancel may fail if job is not actively being processed, but should not return bad request
	// It should return either OK or an error about job not running
	if rr.Code == http.StatusBadRequest {
		t.Errorf("expected non-BadRequest status, got %d", rr.Code)
	}
}
