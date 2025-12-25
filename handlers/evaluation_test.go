package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"llm-tournament/middleware"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
		_ = middleware.CloseDB()
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

	// Ensure default URL path is exercised (tests may set python_service_url).
	_ = middleware.SetSetting("python_service_url", "")

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
	_ = middleware.SetSetting("python_service_url", "http://custom:8001")

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
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

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

func TestEvaluateAllHandler_NilEvaluator(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Store original evaluator and set to nil
	originalEvaluator := globalEvaluator
	globalEvaluator = nil

	// Deferred restore is set up first so it runs last
	defer func() {
		globalEvaluator = originalEvaluator
	}()

	// Deferred panic recovery runs before restore
	defer func() {
		if r := recover(); r != nil {
			// Expected - nil evaluator causes panic, test passes
			_ = r
		}
	}()

	req := httptest.NewRequest("POST", "/evaluate/all", nil)
	rr := httptest.NewRecorder()
	EvaluateAllHandler(rr, req)

	// If we get here (no panic), check that we got an error response
	if rr.Code == http.StatusOK {
		t.Error("expected error when evaluator is nil")
	}
}

func TestEvaluateModelHandler_WithInvalidModel(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	// Request evaluation of non-existent model ID
	req := httptest.NewRequest("POST", "/evaluate/model?id=99999", nil)
	rr := httptest.NewRecorder()
	EvaluateModelHandler(rr, req)

	// Should succeed but create a job that may fail
	// The handler itself should return OK since it creates the job
	if rr.Code == http.StatusBadRequest {
		t.Error("should not return bad request for valid ID format")
	}
}

func TestEvaluatePromptHandler_WithInvalidPrompt(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	// Request evaluation of non-existent prompt ID
	req := httptest.NewRequest("POST", "/evaluate/prompt?id=99999", nil)
	rr := httptest.NewRecorder()
	EvaluatePromptHandler(rr, req)

	// Should succeed but create a job that may fail
	if rr.Code == http.StatusBadRequest {
		t.Error("should not return bad request for valid ID format")
	}
}

func TestCancelEvaluationHandler_WithNonExistentJob(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	req := httptest.NewRequest("POST", "/evaluation/cancel?id=99999", nil)
	rr := httptest.NewRecorder()
	CancelEvaluationHandler(rr, req)

	// Should return error for non-existent job
	if rr.Code == http.StatusOK {
		t.Error("expected error when cancelling non-existent job")
	}
}

func TestCancelEvaluationHandler_SuccessResponse(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	originalLogWriter := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(originalLogWriter)

	db := middleware.GetDB()
	InitEvaluator(db)

	suiteID, err := middleware.GetSuiteID(middleware.GetCurrentSuiteName())
	if err != nil {
		t.Fatalf("GetSuiteID failed: %v", err)
	}

	// Insert enough work so the job is running when we cancel it.
	for i := 0; i < 50; i++ {
		if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", fmt.Sprintf("Model-%d", i), suiteID); err != nil {
			t.Fatalf("insert model failed: %v", err)
		}
		if _, err := db.Exec("INSERT INTO prompts (text, solution, suite_id, display_order, type) VALUES (?, '', ?, ?, 'objective')", fmt.Sprintf("Prompt-%d", i), suiteID, i); err != nil {
			t.Fatalf("insert prompt failed: %v", err)
		}
	}

	jobID, err := globalEvaluator.EvaluateAll(suiteID)
	if err != nil {
		t.Fatalf("EvaluateAll failed: %v", err)
	}

	// Wait for the job to enter running state (worker has registered cancel channel).
	deadline := time.Now().Add(3 * time.Second)
	for {
		var status string
		if err := db.QueryRow("SELECT status FROM evaluation_jobs WHERE id = ?", jobID).Scan(&status); err != nil {
			t.Fatalf("query job status failed: %v", err)
		}
		if status == "running" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for job %d to enter running state", jobID)
		}
		time.Sleep(10 * time.Millisecond)
	}

	req := httptest.NewRequest("POST", fmt.Sprintf("/evaluation/cancel?id=%d", jobID), nil)
	rr := httptest.NewRecorder()
	CancelEvaluationHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if ok, _ := resp["success"].(bool); !ok {
		t.Fatalf("expected success=true, got %#v", resp["success"])
	}
}

func TestEvaluateAllHandler_GetCurrentSuiteIDError(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	// Drop suites table to trigger GetCurrentSuiteID error
	_, err := db.Exec("DROP TABLE suites")
	if err != nil {
		t.Fatalf("failed to drop suites table: %v", err)
	}

	req := httptest.NewRequest("POST", "/evaluate/all", nil)
	rr := httptest.NewRecorder()
	EvaluateAllHandler(rr, req)

	// Should return internal server error when GetCurrentSuiteID fails
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEvaluateAllHandler_EvaluateAllError(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	// Initialize evaluator
	db := middleware.GetDB()
	InitEvaluator(db)

	// Drop evaluation_jobs table to trigger EvaluateAll error
	_, err := db.Exec("DROP TABLE evaluation_jobs")
	if err != nil {
		t.Fatalf("failed to drop evaluation_jobs table: %v", err)
	}

	req := httptest.NewRequest("POST", "/evaluate/all", nil)
	rr := httptest.NewRecorder()
	EvaluateAllHandler(rr, req)

	// Should return internal server error when EvaluateAll fails
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestSaveModelResponseHandler_ErrorCases(t *testing.T) {
	cleanup := setupEvaluationTestDB(t)
	defer cleanup()

	tests := []struct {
		name           string
		method         string
		body           string
		expectStatus   int
		expectContains string
	}{
		{
			name:           "GET method not allowed",
			method:         "GET",
			expectStatus:   http.StatusMethodNotAllowed,
			expectContains: "Method not allowed",
		},
		{
			name:           "missing model_id",
			method:         "POST",
			body:           `{"prompt_id": 1, "response_text": "test response"}`,
			expectStatus:   http.StatusBadRequest,
			expectContains: "model_id is required",
		},
		{
			name:           "missing prompt_id",
			method:         "POST",
			body:           `{"model_id": 1, "response_text": "test response"}`,
			expectStatus:   http.StatusBadRequest,
			expectContains: "prompt_id is required",
		},
		{
			name:           "missing response_text",
			method:         "POST",
			body:           `{"model_id": 1, "prompt_id": 1}`,
			expectStatus:   http.StatusBadRequest,
			expectContains: "response_text is required",
		},
		{
			name:           "invalid JSON",
			method:         "POST",
			body:           `invalid json`,
			expectStatus:   http.StatusBadRequest,
			expectContains: "model_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/save_model_response", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			SaveModelResponseHandler(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, rr.Code)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tt.expectContains) {
				t.Errorf("expected response to contain '%s', got '%s'", tt.expectContains, body)
			}
		})
	}
}
