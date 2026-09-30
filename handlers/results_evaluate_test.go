package handlers

import (
	"errors"
	"llm-tournament/middleware"
	"llm-tournament/testutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

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

func TestEvaluateResult_GET_Request(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

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

	// GET request renders the manual scoring page
	evalReq := httptest.NewRequest("GET", "/evaluate_result?model=GetModel&prompt=0", nil)
	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	if evalRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, evalRR.Code, evalRR.Body.String())
	}
	if !strings.Contains(evalRR.Body.String(), "data-score") {
		t.Error("expected rendered evaluate page to contain score buttons")
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

func TestEvaluateResult_GET_ModelResponseUsesCurrentSuite(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Switch to a second suite so the current suite is not suite 1.
	if err := middleware.SetCurrentSuite("second"); err != nil {
		t.Fatalf("failed to switch suite: %v", err)
	}

	// Prompt and model are created in the current (second) suite.
	prompts := []middleware.Prompt{{Text: "suite two prompt"}}
	_ = middleware.WritePrompts(prompts)

	form := url.Values{}
	form.Add("model", "SuiteTwoModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// Save a model response for the current suite's prompt.
	db := middleware.GetDB()
	var modelID, promptID int
	err := db.QueryRow("SELECT id FROM models WHERE name = ? AND suite_id = (SELECT id FROM suites WHERE is_current = 1)", "SuiteTwoModel").Scan(&modelID)
	if err != nil {
		t.Fatalf("failed to resolve model id: %v", err)
	}
	err = db.QueryRow("SELECT id FROM prompts WHERE suite_id = (SELECT id FROM suites WHERE is_current = 1) ORDER BY display_order LIMIT 1").Scan(&promptID)
	if err != nil {
		t.Fatalf("failed to resolve prompt id: %v", err)
	}
	_, _ = db.Exec("INSERT INTO model_responses (model_id, prompt_id, response_text, response_source) VALUES (?, ?, 'suite two response', 'manual')", modelID, promptID)

	evalReq := httptest.NewRequest("GET", "/evaluate?model=SuiteTwoModel&prompt=0", nil)
	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	if evalRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, evalRR.Code, evalRR.Body.String())
	}
	if !strings.Contains(evalRR.Body.String(), "suite two response") {
		t.Error("expected the saved model response from the current suite to appear in the evaluate page body")
	}
}

func TestEvaluateResult_GET_RendersScoreColorsAndSanitizedMarkdown(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	prompts := []middleware.Prompt{{Text: "check **markdown** color"}}
	_ = middleware.WritePrompts(prompts)

	form := url.Values{}
	form.Add("model", "ColorModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	evalReq := httptest.NewRequest("GET", "/evaluate?model=ColorModel&prompt=0", nil)
	evalRR := httptest.NewRecorder()
	EvaluateResult(evalRR, evalReq)

	if evalRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, evalRR.Code, evalRR.Body.String())
	}
	body := evalRR.Body.String()
	for _, want := range []string{"background-color: #7cff6b", "background-color: #808080", "check <strong>markdown</strong>"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in rendered evaluate body", want)
		}
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

func TestEvaluateResultHandler_RedirectsWhenModelOrPromptMissing(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
		},
		Results: map[string]middleware.Result{
			"Model1": {Scores: []int{100, 80}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	tests := []struct {
		name           string
		url            string
		expectRedirect bool
		expectLocation string
	}{
		{
			name:           "missing both model and prompt",
			url:            "/evaluate",
			expectRedirect: true,
			expectLocation: "/results",
		},
		{
			name:           "missing model",
			url:            "/evaluate?prompt=0",
			expectRedirect: true,
			expectLocation: "/results",
		},
		{
			name:           "missing prompt",
			url:            "/evaluate?model=Model1",
			expectRedirect: true,
			expectLocation: "/results",
		},
		{
			name:           "empty model parameter",
			url:            "/evaluate?model=&prompt=0",
			expectRedirect: true,
			expectLocation: "/results",
		},
		{
			name:           "empty prompt parameter",
			url:            "/evaluate?model=Model1&prompt=",
			expectRedirect: true,
			expectLocation: "/results",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			rr := httptest.NewRecorder()
			handler.EvaluateResultHandler(rr, req)

			if tt.expectRedirect {
				if rr.Code != http.StatusSeeOther {
					t.Errorf("expected redirect status %d, got %d", http.StatusSeeOther, rr.Code)
				}
				location := rr.Header().Get("Location")
				if location != tt.expectLocation {
					t.Errorf("expected redirect to '%s', got '%s'", tt.expectLocation, location)
				}
			}
		})
	}
}
