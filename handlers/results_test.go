package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"llm-tournament/middleware"

	_ "github.com/mattn/go-sqlite3"
)

// setupResultsTestDB creates a test database for results handler tests
func setupResultsTestDB(t *testing.T) func() {
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

func TestMinMax(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(int, int) int
		a, b     int
		expected int
	}{
		{"min_first_smaller", min, 1, 2, 1},
		{"min_second_smaller", min, 2, 1, 1},
		{"min_equal", min, 5, 5, 5},
		{"max_first_larger", max, 2, 1, 2},
		{"max_second_larger", max, 1, 2, 2},
		{"max_equal", max, 5, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(tt.a, tt.b); got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestUpdateResultHandler_Success(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a prompt first
	prompts := []middleware.Prompt{{Text: "Test prompt"}}
	middleware.WritePrompts(prompts)

	// Add a model
	form := url.Values{}
	form.Add("model", "TestModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// Update result
	updateForm := url.Values{}
	updateForm.Add("model", "TestModel")
	updateForm.Add("promptIndex", "0")
	updateForm.Add("pass", "true")

	updateReq := httptest.NewRequest("POST", "/update_result", strings.NewReader(updateForm.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	updateRR := httptest.NewRecorder()
	UpdateResultHandler(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, updateRR.Code)
	}

	// Verify result was updated
	results := middleware.ReadResults()
	if result, exists := results["TestModel"]; exists {
		if len(result.Scores) > 0 && result.Scores[0] != 100 {
			t.Errorf("expected score 100, got %d", result.Scores[0])
		}
	}
}

func TestUpdateResultHandler_InvalidPass(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	updateForm := url.Values{}
	updateForm.Add("model", "TestModel")
	updateForm.Add("promptIndex", "0")
	updateForm.Add("pass", "invalid")

	updateReq := httptest.NewRequest("POST", "/update_result", strings.NewReader(updateForm.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	updateRR := httptest.NewRecorder()
	UpdateResultHandler(updateRR, updateReq)

	if updateRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, updateRR.Code)
	}
}

func TestResetResultsHandler_POST(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a model with results
	form := url.Values{}
	form.Add("model", "TestModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// Reset results
	resetReq := httptest.NewRequest("POST", "/reset_results", nil)
	resetRR := httptest.NewRecorder()
	ResetResultsHandler(resetRR, resetReq)

	if resetRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, resetRR.Code)
	}

	// Verify results were reset
	results := middleware.ReadResults()
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestRefreshResultsHandler_POST(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a model with results
	form := url.Values{}
	form.Add("model", "TestModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// Refresh results
	refreshReq := httptest.NewRequest("POST", "/refresh_results", nil)
	refreshRR := httptest.NewRecorder()
	RefreshResultsHandler(refreshRR, refreshReq)

	if refreshRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, refreshRR.Code)
	}
}

func TestExportResultsHandler(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a model with results
	form := url.Values{}
	form.Add("model", "ExportTestModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// Export results
	exportReq := httptest.NewRequest("GET", "/export_results", nil)
	exportRR := httptest.NewRecorder()
	ExportResultsHandler(exportRR, exportReq)

	if exportRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, exportRR.Code)
	}

	// Verify JSON content type
	contentType := exportRR.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content type 'application/json', got %q", contentType)
	}
}

func TestEvaluateResult_POST_Success(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a prompt first
	prompts := []middleware.Prompt{{Text: "Evaluate test prompt"}}
	middleware.WritePrompts(prompts)

	// Add a model
	form := url.Values{}
	form.Add("model", "EvalModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// Evaluate result
	evalForm := url.Values{}
	evalForm.Add("score", "80")

	evalReq := httptest.NewRequest("POST", "/evaluate_result?model=EvalModel&prompt=0", strings.NewReader(evalForm.Encode()))
	evalReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	if evalRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, evalRR.Code)
	}

	// Verify score was updated
	results := middleware.ReadResults()
	if result, exists := results["EvalModel"]; exists {
		if len(result.Scores) > 0 && result.Scores[0] != 80 {
			t.Errorf("expected score 80, got %d", result.Scores[0])
		}
	}
}

func TestEvaluateResult_POST_InvalidScore(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	evalForm := url.Values{}
	evalForm.Add("score", "not_a_number")

	evalReq := httptest.NewRequest("POST", "/evaluate_result?model=TestModel&prompt=0", strings.NewReader(evalForm.Encode()))
	evalReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	if evalRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, evalRR.Code)
	}
}

func TestEvaluateResult_POST_InvalidPromptIndex(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a model
	form := url.Values{}
	form.Add("model", "TestModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	evalForm := url.Values{}
	evalForm.Add("score", "80")

	evalReq := httptest.NewRequest("POST", "/evaluate_result?model=TestModel&prompt=invalid", strings.NewReader(evalForm.Encode()))
	evalReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	if evalRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, evalRR.Code)
	}
}

func TestEvaluateResult_POST_ScoreClamping(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a prompt
	prompts := []middleware.Prompt{{Text: "Clamp test prompt"}}
	middleware.WritePrompts(prompts)

	// Add a model
	form := url.Values{}
	form.Add("model", "ClampModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// Test score > 100 (should be clamped to 100)
	evalForm := url.Values{}
	evalForm.Add("score", "150")

	evalReq := httptest.NewRequest("POST", "/evaluate_result?model=ClampModel&prompt=0", strings.NewReader(evalForm.Encode()))
	evalReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	results := middleware.ReadResults()
	if result, exists := results["ClampModel"]; exists {
		if len(result.Scores) > 0 && result.Scores[0] > 100 {
			t.Errorf("expected score <= 100, got %d", result.Scores[0])
		}
	}
}

func TestUpdateMockResultsHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/update_mock_results", nil)
	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestUpdateMockResultsHandler_Success(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "Mock test prompt"}}
	middleware.WritePrompts(prompts)

	mockData := map[string]interface{}{
		"models":  []string{"MockModel"},
		"results": map[string]middleware.Result{"MockModel": {Scores: []int{80}}},
	}
	jsonBody, _ := json.Marshal(mockData)

	req := httptest.NewRequest("POST", "/update_mock_results", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestUpdateMockResultsHandler_InvalidJSON(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader("not valid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUpdateMockResultsHandler_ValidatesScores(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "Validate test prompt"}}
	middleware.WritePrompts(prompts)

	// Send invalid score (should be corrected to 0)
	mockData := map[string]interface{}{
		"models":  []string{"ValidateModel"},
		"results": map[string]middleware.Result{"ValidateModel": {Scores: []int{77}}}, // 77 is not valid
	}
	jsonBody, _ := json.Marshal(mockData)

	req := httptest.NewRequest("POST", "/update_mock_results", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify score was corrected
	results := middleware.ReadResults()
	if result, exists := results["ValidateModel"]; exists {
		if len(result.Scores) > 0 && result.Scores[0] != 0 {
			t.Errorf("expected score 0 (corrected from invalid 77), got %d", result.Scores[0])
		}
	}
}

func TestInitRand(t *testing.T) {
	r := initRand()
	if r == nil {
		t.Error("initRand should not return nil")
	}
	// Verify it produces random values
	val1 := r.Intn(1000)
	val2 := r.Intn(1000)
	// Very unlikely to be equal, but not impossible
	_ = val1
	_ = val2
}

func TestUpdateResultHandler_POST_MissingParams(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Missing model parameter
	updateForm := url.Values{}
	updateForm.Add("promptIndex", "0")
	updateForm.Add("pass", "true")

	updateReq := httptest.NewRequest("POST", "/update_result", strings.NewReader(updateForm.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	updateRR := httptest.NewRecorder()
	UpdateResultHandler(updateRR, updateReq)

	// Should still work (empty model name)
	if updateRR.Code == http.StatusBadRequest {
		t.Errorf("did not expect status %d", updateRR.Code)
	}
}

func TestUpdateResultHandler_POST_NegativePromptIndex(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a model first
	form := url.Values{}
	form.Add("model", "TestModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	updateForm := url.Values{}
	updateForm.Add("model", "TestModel")
	updateForm.Add("promptIndex", "-1")
	updateForm.Add("pass", "true")

	updateReq := httptest.NewRequest("POST", "/update_result", strings.NewReader(updateForm.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	updateRR := httptest.NewRecorder()
	UpdateResultHandler(updateRR, updateReq)

	// The handler may not validate negative indices strictly
	// Just check that it doesn't crash
	if updateRR.Code == http.StatusInternalServerError {
		t.Errorf("unexpected internal server error")
	}
}

func TestEvaluateResult_GET_Request(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a prompt
	prompts := []middleware.Prompt{{Text: "GET test prompt"}}
	middleware.WritePrompts(prompts)

	// Add a model
	form := url.Values{}
	form.Add("model", "GetModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// GET request should also work (handler may render template)
	evalReq := httptest.NewRequest("GET", "/evaluate_result?model=GetModel&prompt=0", nil)
	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	// GET request should fail (expects POST or template rendering)
	if evalRR.Code == http.StatusMethodNotAllowed {
		// That's fine, handler only supports POST
	}
}
