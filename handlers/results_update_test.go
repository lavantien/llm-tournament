package handlers

import (
	"errors"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

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

	// Missing model must be rejected without touching stored results
	if updateRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for missing model, got %d", http.StatusBadRequest, updateRR.Code)
	}
	if results := middleware.ReadResults(); len(results) != 0 {
		t.Errorf("expected no results to be written, got %d models", len(results))
	}
}

func TestUpdateResultHandler_POST_NegativePromptIndex(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Seed one prompt and one scored model so rejection leaves state untouched.
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Bounds prompt"}})
	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"TestModel": {Scores: []int{40}},
	})

	updateForm := url.Values{}
	updateForm.Add("model", "TestModel")
	updateForm.Add("promptIndex", "-1")
	updateForm.Add("pass", "true")

	updateReq := httptest.NewRequest("POST", "/update_result", strings.NewReader(updateForm.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	updateRR := httptest.NewRecorder()
	UpdateResultHandler(updateRR, updateReq)

	// The bounds check must reject negative indices with 400, not write cell 0.
	if updateRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for negative promptIndex, got %d", http.StatusBadRequest, updateRR.Code)
	}
	results := middleware.ReadResults()
	if result, exists := results["TestModel"]; exists {
		if len(result.Scores) != 1 || result.Scores[0] != 40 {
			t.Errorf("expected score to stay 40 after rejected update, got %v", result.Scores)
		}
	} else {
		t.Error("expected results for TestModel to still exist")
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
