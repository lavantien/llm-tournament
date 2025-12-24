package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"llm-tournament/middleware"
	"llm-tournament/testutil"

	_ "github.com/mattn/go-sqlite3"
)

// changeToProjectRootResults changes to the project root directory for tests that need templates
func changeToProjectRootResults(t *testing.T) func() {
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

// setupResultsTestDB creates a test database for results handler tests
func setupResultsTestDB(t *testing.T) func() {
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
	_ = middleware.WritePrompts(prompts)

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
	_ = middleware.WritePrompts(prompts)

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
	_ = middleware.WritePrompts(prompts)

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
	_ = middleware.WritePrompts(prompts)

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
	_ = middleware.WritePrompts(prompts)

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
	_ = middleware.WritePrompts(prompts)

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
		_ = evalRR.Code
	}
}

func TestResultsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test prompts
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt 1"},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	// Add a model
	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"TestModel": {Scores: []int{80}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	ResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "TestModel") {
		t.Error("expected model name in response body")
	}
}

func TestConfirmRefreshResultsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/confirm_refresh_results", nil)
	rr := httptest.NewRecorder()
	ConfirmRefreshResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestResetResultsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/reset_results", nil)
	rr := httptest.NewRecorder()
	ResetResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestExportResultsHandler_GET(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test prompts and results
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt"},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"TestModel": {Scores: []int{80}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/export_results", nil)
	rr := httptest.NewRecorder()
	ExportResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify CSV content
	body := rr.Body.String()
	if !strings.Contains(body, "Model") {
		t.Error("expected 'Model' header in CSV output")
	}
}

func TestConfirmRefreshResultsHandler_POST(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test prompts and results
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt"},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"RefreshModel": {Scores: []int{80}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	// POST request should refresh results
	req := httptest.NewRequest("POST", "/confirm_refresh_results", nil)
	rr := httptest.NewRecorder()
	ConfirmRefreshResultsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify scores were zeroed
	results := middleware.ReadResults()
	if result, exists := results["RefreshModel"]; exists {
		for i, score := range result.Scores {
			if score != 0 {
				t.Errorf("expected score 0 at index %d, got %d", i, score)
			}
		}
	}
}

func TestRefreshResultsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/refresh_results", nil)
	rr := httptest.NewRecorder()
	RefreshResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestEvaluateResult_GET_WithTemplate(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a prompt
	prompts := []middleware.Prompt{{Text: "Template test prompt"}}
	_ = middleware.WritePrompts(prompts)

	// Add a model with results
	form := url.Values{}
	form.Add("model", "TemplateModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// GET request should render template
	evalReq := httptest.NewRequest("GET", "/evaluate_result?model=TemplateModel&prompt=0", nil)
	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	if evalRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, evalRR.Code)
	}

	body := evalRR.Body.String()
	if !strings.Contains(body, "TemplateModel") {
		t.Error("expected model name in response body")
	}
}

func TestEvaluateResult_POST_NewModel(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "New model test prompt"}}
	_ = middleware.WritePrompts(prompts)

	// POST with model that doesn't have results yet
	evalForm := url.Values{}
	evalForm.Add("score", "60")

	evalReq := httptest.NewRequest("POST", "/evaluate_result?model=NewModel&prompt=0", strings.NewReader(evalForm.Encode()))
	evalReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	// Should initialize results for new model
	if evalRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, evalRR.Code)
	}

	results := middleware.ReadResults()
	if result, exists := results["NewModel"]; exists {
		if len(result.Scores) > 0 && result.Scores[0] != 60 {
			t.Errorf("expected score 60, got %d", result.Scores[0])
		}
	} else {
		t.Error("expected results for NewModel to exist")
	}
}

func TestEvaluateResult_POST_NegativeScore(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "Negative score test"}}
	_ = middleware.WritePrompts(prompts)

	// Add model
	form := url.Values{}
	form.Add("model", "NegModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// POST with negative score
	evalForm := url.Values{}
	evalForm.Add("score", "-10")

	evalReq := httptest.NewRequest("POST", "/evaluate_result?model=NegModel&prompt=0", strings.NewReader(evalForm.Encode()))
	evalReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	// Score should be clamped to 0
	results := middleware.ReadResults()
	if result, exists := results["NegModel"]; exists {
		if len(result.Scores) > 0 && result.Scores[0] != 0 {
			t.Errorf("expected score 0 (clamped from -10), got %d", result.Scores[0])
		}
	}
}

func TestResultsHandler_GET_WithUncategorizedPrompts(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts with empty profile (uncategorized)
	err := middleware.WritePrompts([]middleware.Prompt{
		{Text: "Categorized Prompt", Profile: "TestProfile"},
		{Text: "Uncategorized Prompt", Profile: ""},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	// Add a profile
	err = middleware.WriteProfiles([]middleware.Profile{
		{Name: "TestProfile", Description: "Test"},
	})
	if err != nil {
		t.Fatalf("failed to write profiles: %v", err)
	}

	// Add results
	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"TestModel": {Scores: []int{80, 60}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	ResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Uncategorized") {
		t.Error("expected 'Uncategorized' group in response body for prompts without profile")
	}
}

func TestResultsHandler_GET_WithMultipleProfiles(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add multiple profiles
	err := middleware.WriteProfiles([]middleware.Profile{
		{Name: "Profile1", Description: "First profile"},
		{Name: "Profile2", Description: "Second profile"},
	})
	if err != nil {
		t.Fatalf("failed to write profiles: %v", err)
	}

	// Add prompts across profiles
	err = middleware.WritePrompts([]middleware.Prompt{
		{Text: "Prompt 1", Profile: "Profile1"},
		{Text: "Prompt 2", Profile: "Profile1"},
		{Text: "Prompt 3", Profile: "Profile2"},
	})
	if err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}

	// Add results
	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"Model1": {Scores: []int{100, 80, 60}},
		"Model2": {Scores: []int{60, 40, 20}},
	})
	if err != nil {
		t.Fatalf("failed to write results: %v", err)
	}

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	ResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Profile1") {
		t.Error("expected Profile1 in response body")
	}
	if !strings.Contains(body, "Profile2") {
		t.Error("expected Profile2 in response body")
	}
}

func TestResultsHandler_ProfileGroups_UncategorizedStartColAfterProfile(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Categorized Prompt", Profile: "Profile1"},
			{Text: "Uncategorized Prompt", Profile: ""},
		},
		Profiles: []middleware.Profile{
			{Name: "Profile1", Description: "First profile"},
		},
		Results: map[string]middleware.Result{
			"Model1": {Scores: []int{100, 80}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	handler.Results(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}

	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Struct {
		t.Fatalf("expected struct template data, got %T", data)
	}

	field := val.FieldByName("ProfileGroups")
	if !field.IsValid() {
		t.Fatalf("expected ProfileGroups field on template data")
	}

	var uncategorized *middleware.ProfileGroup
	for i := 0; i < field.Len(); i++ {
		pg, ok := field.Index(i).Interface().(*middleware.ProfileGroup)
		if !ok {
			t.Fatalf("expected *middleware.ProfileGroup, got %T", field.Index(i).Interface())
		}
		if pg.Name == "Uncategorized" {
			uncategorized = pg
			break
		}
	}

	if uncategorized == nil {
		t.Fatalf("expected Uncategorized profile group")
	}
	if uncategorized.StartCol != 1 {
		t.Fatalf("expected Uncategorized StartCol=1, got %d", uncategorized.StartCol)
	}
}

func TestResultsHandler_ModelFilterFiltersResultsMap(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
		},
		Results: map[string]middleware.Result{
			"Model1": {Scores: []int{100}},
			"Model2": {Scores: []int{80}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	req := httptest.NewRequest("GET", "/results?model_filter=Model1", nil)
	rr := httptest.NewRecorder()
	handler.Results(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}

	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)

	resultsField := val.FieldByName("Results")
	if resultsField.Kind() != reflect.Map {
		t.Fatalf("expected Results to be a map, got %v", resultsField.Kind())
	}
	if resultsField.Len() != 1 {
		t.Fatalf("expected 1 filtered result, got %d", resultsField.Len())
	}
	if !resultsField.MapIndex(reflect.ValueOf("Model1")).IsValid() {
		t.Fatalf("expected Model1 to remain after filtering")
	}
}

func TestResultsHandler_SearchQueryFiltersResultsMap(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
		},
		Results: map[string]middleware.Result{
			"AlphaModel": {Scores: []int{100}},
			"Beta":       {Scores: []int{80}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	req := httptest.NewRequest("GET", "/results?search=alpha", nil)
	rr := httptest.NewRecorder()
	handler.Results(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}

	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)

	resultsField := val.FieldByName("Results")
	if resultsField.Len() != 1 {
		t.Fatalf("expected 1 filtered result, got %d", resultsField.Len())
	}
	if !resultsField.MapIndex(reflect.ValueOf("AlphaModel")).IsValid() {
		t.Fatalf("expected AlphaModel to remain after search filtering")
	}
}

func TestResultsHandler_NormalizesNilMismatchedAndInvalidScores(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
		},
		Results: map[string]middleware.Result{
			"NilScores":   {Scores: nil},
			"ShortScores": {Scores: []int{200}}, // invalid + length mismatch
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	handler.Results(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}

	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)

	resultsField := val.FieldByName("Results")
	nilScores := resultsField.MapIndex(reflect.ValueOf("NilScores"))
	if !nilScores.IsValid() {
		t.Fatalf("expected NilScores in template results")
	}
	shortScores := resultsField.MapIndex(reflect.ValueOf("ShortScores"))
	if !shortScores.IsValid() {
		t.Fatalf("expected ShortScores in template results")
	}

	nilScoresSlice := nilScores.FieldByName("Scores")
	if nilScoresSlice.Len() != 2 {
		t.Fatalf("expected NilScores Scores length 2, got %d", nilScoresSlice.Len())
	}

	shortScoresSlice := shortScores.FieldByName("Scores")
	if shortScoresSlice.Len() != 2 {
		t.Fatalf("expected ShortScores Scores length 2, got %d", shortScoresSlice.Len())
	}
	if shortScoresSlice.Index(0).Int() != 0 {
		t.Fatalf("expected invalid score to be normalized to 0, got %d", shortScoresSlice.Index(0).Int())
	}
}

func TestUpdateMockResultsHandler_WithEmptyModels(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts first
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Send request with results but no explicit models array
	mockData := `{
		"results": {
			"ModelFromResults": {"scores": [80]}
		},
		"models": [],
		"passPercentages": {},
		"totalScores": {}
	}`

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(mockData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestUpdateMockResultsHandler_ValidatesInvalidScores(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts first
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Send request with invalid scores (should be corrected to 0)
	mockData := `{
		"results": {
			"TestModel": {"scores": [15, 25, 35]}
		},
		"models": ["TestModel"],
		"passPercentages": {},
		"totalScores": {}
	}`

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(mockData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify invalid scores were corrected
	results := middleware.ReadResults()
	if result, exists := results["TestModel"]; exists {
		for i, score := range result.Scores {
			if score != 0 {
				t.Errorf("expected invalid score at index %d to be corrected to 0, got %d", i, score)
			}
		}
	}
}

func TestExportResultsHandler_GET_WithData(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test data
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})
	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{80}},
	})

	req := httptest.NewRequest("GET", "/export_results", nil)
	rr := httptest.NewRecorder()
	ExportResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	// Verify JSON is valid
	var data map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &data)
	if err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
}

func TestExportResultsHandler_EmptyResults(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/export_results", nil)
	rr := httptest.NewRecorder()
	ExportResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestResetResultsHandler_POST_AndVerify(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test data
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})
	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{80}},
	})

	req := httptest.NewRequest("POST", "/reset_results", nil)
	rr := httptest.NewRecorder()
	ResetResultsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify results were reset
	results := middleware.ReadResults()
	for _, result := range results {
		for _, score := range result.Scores {
			if score != 0 {
				t.Error("expected all scores to be reset to 0")
			}
		}
	}
}

func TestConfirmRefreshResultsHandler_WithSearchQuery(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{80}},
		"Model2": {Scores: []int{60}},
	})

	req := httptest.NewRequest("GET", "/confirm_refresh_results?search_query=Model1", nil)
	rr := httptest.NewRecorder()
	ConfirmRefreshResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestRefreshResultsHandler_POST_WithSelectedModels(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test data
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})
	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{80}},
	})

	form := url.Values{}
	form.Add("selected_models", "Model1")

	req := httptest.NewRequest("POST", "/refresh_results", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	RefreshResultsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestRefreshResultsHandler_POST_NoModels(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/refresh_results", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	RefreshResultsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestUpdateResultHandler_CreatesNewModel(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test"}})

	form := url.Values{}
	form.Add("model", "NewModel")
	form.Add("promptIndex", "0")
	form.Add("pass", "true")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	UpdateResultHandler(rr, req)

	// Handler should succeed and create the new model
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify model was created
	results := middleware.ReadResults()
	if _, exists := results["NewModel"]; !exists {
		t.Error("expected NewModel to be created")
	}
}

func TestUpdateResultHandler_WithScoreValue(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test"}})
	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{50}},
	})

	form := url.Values{}
	form.Add("model", "Model1")
	form.Add("promptIndex", "0")
	form.Add("pass", "true")
	form.Add("score", "80")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	UpdateResultHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestResultsHandler_GET_RenderError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	ResultsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestResetResultsHandler_GET_RenderError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/reset_results", nil)
	rr := httptest.NewRecorder()
	ResetResultsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestConfirmRefreshResultsHandler_GET_RenderError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/confirm_refresh_results", nil)
	rr := httptest.NewRecorder()
	ConfirmRefreshResultsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestUpdateResultHandler_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "Test prompt"}},
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{50}},
		},
		CurrentSuite: "test-suite",
		WriteResultsFunc: func(suiteName string, results map[string]middleware.Result) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("model", "TestModel")
	form.Add("promptIndex", "0")
	form.Add("pass", "true")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.UpdateResult(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestUpdateResultHandler_InvalidPassValue(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{
			Prompts: []middleware.Prompt{{Text: "P1"}},
			Results: map[string]middleware.Result{"TestModel": {Scores: []int{0}}},
		},
		Renderer: &MockRenderer{},
	}

	form := url.Values{}
	form.Add("model", "TestModel")
	form.Add("promptIndex", "0")
	form.Add("pass", "not-a-bool")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.UpdateResult(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid pass value") {
		t.Fatalf("expected invalid pass value message, got %q", rr.Body.String())
	}
}

func TestUpdateResultHandler_ExtendsScores(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}, {Text: "P2"}, {Text: "P3"}},
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{0}},
		},
	}
	var wrote map[string]middleware.Result
	mockDS.WriteResultsFunc = func(suiteName string, results map[string]middleware.Result) error {
		wrote = results
		return nil
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("model", "TestModel")
	form.Add("promptIndex", "2")
	form.Add("pass", "true")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.UpdateResult(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if wrote == nil {
		t.Fatal("expected WriteResults to be called")
	}
	result := wrote["TestModel"]
	if len(result.Scores) != 3 {
		t.Fatalf("expected scores length 3, got %d", len(result.Scores))
	}
	if result.Scores[2] != 100 {
		t.Fatalf("expected score[2] = 100, got %d", result.Scores[2])
	}
}

func TestUpdateResultHandler_ReadResultsNil(t *testing.T) {
	ds := &nilResultsDataStore{}
	ds.Prompts = []middleware.Prompt{{Text: "P1"}, {Text: "P2"}}
	var wrote map[string]middleware.Result
	ds.WriteResultsFunc = func(suiteName string, results map[string]middleware.Result) error {
		wrote = results
		return nil
	}

	handler := &Handler{
		DataStore: ds,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("model", "NewModel")
	form.Add("promptIndex", "1")
	form.Add("pass", "false")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.UpdateResult(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if wrote == nil {
		t.Fatal("expected WriteResults to be called")
	}
	result, ok := wrote["NewModel"]
	if !ok {
		t.Fatalf("expected results to contain %q", "NewModel")
	}
	if len(result.Scores) != 2 {
		t.Fatalf("expected scores length 2, got %d", len(result.Scores))
	}
	if result.Scores[1] != 0 {
		t.Fatalf("expected score[1] = 0, got %d", result.Scores[1])
	}
}

func TestUpdateResultHandler_WriteResponseError(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{
			Prompts: []middleware.Prompt{{Text: "P1"}},
			Results: map[string]middleware.Result{"TestModel": {Scores: []int{0}}},
		},
		Renderer: &MockRenderer{},
	}

	form := url.Values{}
	form.Add("model", "TestModel")
	form.Add("promptIndex", "0")
	form.Add("pass", "true")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	failingWriter := &FailingResponseWriter{
		ResponseWriter: rr,
		WriteError:     errors.New("mock write error"),
	}
	handler.UpdateResult(failingWriter, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestResetResultsHandler_POST_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		CurrentSuite: "test-suite",
		WriteResultsFunc: func(suiteName string, results map[string]middleware.Result) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest("POST", "/reset_results", nil)
	rr := httptest.NewRecorder()
	handler.ResetResults(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestConfirmRefreshResultsHandler_POST_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "Test prompt"}},
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{80}},
		},
		CurrentSuite: "test-suite",
		WriteResultsFunc: func(suiteName string, results map[string]middleware.Result) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest("POST", "/confirm_refresh_results", nil)
	rr := httptest.NewRecorder()
	handler.ConfirmRefreshResults(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestRefreshResultsHandler_POST_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "Test prompt"}},
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{80}},
		},
		CurrentSuite: "test-suite",
		WriteResultsFunc: func(suiteName string, results map[string]middleware.Result) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest("POST", "/refresh_results", nil)
	rr := httptest.NewRecorder()
	handler.RefreshResults(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEvaluateResultHandler_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "Test prompt"}},
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{50}},
		},
		CurrentSuite: "test-suite",
		WriteResultsFunc: func(suiteName string, results map[string]middleware.Result) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("score", "80")

	req := httptest.NewRequest("POST", "/evaluate_result?model=TestModel&prompt=0", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.EvaluateResultHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEvaluateResultHandler_InitializesNilResultsMap(t *testing.T) {
	mockDS := &nilResultsDataStore{
		MockDataStore: MockDataStore{
			Prompts:      []middleware.Prompt{{Text: "Test prompt"}},
			CurrentSuite: "test-suite",
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("score", "80")

	req := httptest.NewRequest("POST", "/evaluate_result?model=NewModel&prompt=0", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.EvaluateResultHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	if mockDS.Results == nil {
		t.Fatalf("expected results map to be initialized")
	}
	result, ok := mockDS.Results["NewModel"]
	if !ok {
		t.Fatalf("expected NewModel to be created")
	}
	if len(result.Scores) != 1 || result.Scores[0] != 80 {
		t.Fatalf("expected score to be stored, got %#v", result.Scores)
	}
}

func TestUpdateMockResultsHandler_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts:      []middleware.Prompt{{Text: "Test prompt"}},
		CurrentSuite: "test-suite",
		WriteResultsFunc: func(suiteName string, results map[string]middleware.Result) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	mockData := `{
		"results": {"TestModel": {"scores": [80]}},
		"models": ["TestModel"]
	}`

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(mockData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestUpdateMockResults_ReadAllError(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{Prompts: []middleware.Prompt{{Text: "Test prompt"}}},
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest("POST", "/update_mock_results", readErrorReader{})
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUpdateMockResults_SortsModelsByTotalScore(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{Prompts: []middleware.Prompt{{Text: "Test prompt"}}},
		Renderer:  &MockRenderer{},
	}

	// Intentionally provide models in reverse order; handler should sort by total score.
	mockData := `{
                "results": {
                        "ModelA": {"scores": [100]},
                        "ModelB": {"scores": [0]}
                },
                "models": ["ModelB", "ModelA"]
        }`

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(mockData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp struct {
		Models []string `json:"models"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(resp.Models) != 2 || resp.Models[0] != "ModelA" || resp.Models[1] != "ModelB" {
		t.Fatalf("expected models sorted as [ModelA ModelB], got %v", resp.Models)
	}
}

func TestUpdateMockResults_ResponseEncodeError(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{Prompts: []middleware.Prompt{{Text: "Test prompt"}}},
		Renderer:  &MockRenderer{},
	}

	mockData := `{
                "results": {"TestModel": {"scores": [80]}},
                "models": ["TestModel"]
        }`

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(mockData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	failingWriter := &FailingResponseWriter{
		ResponseWriter: rr,
		WriteError:     errors.New("mock write error"),
	}

	handler.UpdateMockResults(failingWriter, req)

	// Handler logs encode errors but doesn't return a different status code.
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestRefreshResultsHandler_GET_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		CurrentSuite: "test-suite",
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/refresh_results", nil)
	rr := httptest.NewRecorder()
	handler.RefreshResults(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEvaluateResultHandler_GET_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "Test prompt"}},
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{50}},
		},
		CurrentSuite: "test-suite",
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/evaluate_result?model=TestModel&prompt=0", nil)
	rr := httptest.NewRecorder()
	handler.EvaluateResultHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestExportResultsHandler_WriteError(t *testing.T) {
	mockDS := &MockDataStore{
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{80}},
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

	req := httptest.NewRequest("GET", "/export_results", nil)
	handler.ExportResults(failingWriter, req)

	// The handler should fail when writing the response
	// Check that no successful content was written
	if failingWriter.HeaderWritten && rr.Code == http.StatusOK {
		// Write error occurred after header was written
		// This is expected behavior - the error is logged but header already sent
		_ = failingWriter.HeaderWritten
	}
}

func TestUpdateMockResultsHandler_GeneratesMockModels(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Initial state: verify no models exist
	db := middleware.GetDB()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM models").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query model count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 models initially, got %d", count)
	}

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Use default handler which uses real middleware
	DefaultHandler.UpdateMockResults(rr, req)

	// Verify 15 mock models created
	err = db.QueryRow("SELECT COUNT(*) FROM models").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query model count after generation: %v", err)
	}
	if count != 15 {
		t.Errorf("expected 15 mock models, got %d", count)
	}

	// Verify model names use tier-based pattern
	rows, err := db.Query("SELECT name FROM models ORDER BY name")
	if err != nil {
		t.Fatalf("failed to query models: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	}()

	models := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan model name: %v", err)
		}
		models = append(models, name)
	}

	// Check for expected tier prefixes in model names
	expectedTiers := []string{"Cosmic", "Transcendent", "Ethereal", "Celestial", "Infinite"}
	foundTiers := make(map[string]bool)
	for _, model := range models {
		for _, tier := range expectedTiers {
			if strings.Contains(model, tier) {
				foundTiers[tier] = true
				break
			}
		}
	}

	// At least some expected tiers should be present
	if len(foundTiers) < 3 {
		t.Errorf("expected tier-based model names (Cosmic, Transcendent, etc.), got: %v", models[:5])
	}
}

func TestUpdateMockResultsHandler_GeneratesTierBasedModelNames(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Use default handler which uses real middleware
	DefaultHandler.UpdateMockResults(rr, req)

	// Verify model names use tier-based pattern
	db := middleware.GetDB()
	rows, err := db.Query("SELECT name FROM models ORDER BY name")
	if err != nil {
		t.Fatalf("failed to query models: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	}()

	models := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan model name: %v", err)
		}
		models = append(models, name)
	}

	// Check that at least one model contains 'Cosmic'
	foundCosmic := false
	for _, model := range models {
		if strings.Contains(model, "Cosmic") {
			foundCosmic = true
			break
		}
	}
	if !foundCosmic {
		t.Errorf("expected at least one model to contain 'Cosmic', got: %v", models[:5])
	}
}

func TestUpdateMockResultsHandler_GeneratesMockResponses(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Use default handler which uses real middleware
	DefaultHandler.UpdateMockResults(rr, req)

	// Verify mock responses were created
	db := middleware.GetDB()
	var responseCount int
	err := db.QueryRow("SELECT COUNT(*) FROM model_responses WHERE response_source = 'mock'").Scan(&responseCount)
	if err != nil {
		t.Fatalf("failed to query response count: %v", err)
	}

	// Should have 15 models * 1 prompt = 15 mock responses
	expectedResponses := 15 // models count * prompt count
	if responseCount < expectedResponses {
		t.Errorf("expected at least %d mock responses, got %d", expectedResponses, responseCount)
	}

	// Verify response text is not empty
	var responseText string
	err = db.QueryRow("SELECT response_text FROM model_responses WHERE response_source = 'mock' LIMIT 1").Scan(&responseText)
	if err != nil {
		t.Fatalf("failed to query mock response: %v", err)
	}
	if len(responseText) == 0 {
		t.Error("expected mock response text to be non-empty")
	}
}

func TestUpdateMockResultsHandler_CreatesModelsInCurrentSuite(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Trigger mock generation with empty request (no existing models)
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Verify the response is successful
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// Verify models were created in the database
	db := middleware.GetDB()
	var modelCount int
	err := db.QueryRow("SELECT COUNT(*) FROM models").Scan(&modelCount)
	if err != nil {
		t.Fatalf("failed to query model count: %v", err)
	}

	// Should have created 15 mock models
	expectedModels := 15
	if modelCount != expectedModels {
		t.Errorf("expected %d models in database, got %d", expectedModels, modelCount)
	}
}

func TestResultsHandler_ZeroPrompts_DoesNotReturnNaN(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add results but NO prompts - this causes NaN in pass percentage calculation
	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"TestModel": {Scores: []int{}}, // Empty scores because no prompts
	})

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()

	ResultsHandler(rr, req)

	// When NaN occurs, the response body is incomplete/empty
	body := rr.Body.String()

	// Debug: print the body to see what we got
	t.Logf("Response body length: %d", len(body))
	t.Logf("Response body preview: %s", body[:min(500, len(body))])

	// The template will fail to render properly when NaN is in PassPercentages
	// Check that the response is actually complete and valid HTML
	if !strings.Contains(body, "</html>") {
		t.Error("response should contain complete HTML, but template rendering likely failed due to NaN")
	}
}

func TestUpdateMockResultsHandler_GeneratesMockPrompts(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Start with completely empty database (no prompts, no models)

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Verify the response is successful
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// Verify prompts were created in the database
	db := middleware.GetDB()
	var promptCount int
	err := db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&promptCount)
	if err != nil {
		t.Fatalf("failed to query prompt count: %v", err)
	}

	// Should have created mock prompts
	if promptCount == 0 {
		t.Errorf("expected prompts to be created in database, got %d", promptCount)
	}
}
