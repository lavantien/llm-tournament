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

func TestConfirmRefreshResultsHandler_POST_WithSelectedModels(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test data
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})
	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{80}},
	})

	form := url.Values{}
	form.Add("selected_models", "Model1")

	req := httptest.NewRequest("POST", "/confirm_refresh_results", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	ConfirmRefreshResultsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestConfirmRefreshResultsHandler_POST_NoModels(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/confirm_refresh_results", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	ConfirmRefreshResultsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
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

func TestResetResultsHandler_MethodNotAllowed(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
		},
		Results: map[string]middleware.Result{
			"Model1": {Scores: []int{100}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	// Test PUT method (not GET or POST)
	req := httptest.NewRequest("PUT", "/reset_results", nil)
	rr := httptest.NewRecorder()
	handler.ResetResults(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestConfirmRefreshResultsHandler_MethodNotAllowed(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
		},
		Results: map[string]middleware.Result{
			"Model1": {Scores: []int{100}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	// Test PUT method (not GET or POST)
	req := httptest.NewRequest("PUT", "/confirm_refresh_results", nil)
	rr := httptest.NewRecorder()
	handler.ConfirmRefreshResults(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
