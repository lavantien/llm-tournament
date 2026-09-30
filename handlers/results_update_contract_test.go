package handlers

import (
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// seedUpdateResultFixture writes one prompt and one scored model so contract
// tests can prove the handler leaves stored state untouched on bad input.
func seedUpdateResultFixture(t *testing.T) {
	t.Helper()
	if err := middleware.WritePrompts([]middleware.Prompt{{Text: "Contract prompt"}}); err != nil {
		t.Fatalf("failed to write prompt: %v", err)
	}
	if err := middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{80}},
	}); err != nil {
		t.Fatalf("failed to write results: %v", err)
	}
}

func assertScoreUnchanged(t *testing.T, model string, want int) {
	t.Helper()
	results := middleware.ReadResults()
	result, exists := results[model]
	if !exists {
		t.Fatalf("expected results for %s to still exist", model)
	}
	if len(result.Scores) != 1 || result.Scores[0] != want {
		t.Fatalf("expected stored score for %s to be unchanged at %d, got %v", model, want, result.Scores)
	}
}

func TestUpdateResultContract_GETRejectedWithoutStateChange(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()
	seedUpdateResultFixture(t)

	req := httptest.NewRequest("GET", "/update_result?model=Model1&promptIndex=0&pass=true", nil)
	rr := httptest.NewRecorder()
	UpdateResultHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d for GET, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Method not allowed") {
		t.Fatalf("expected method not allowed message, got %q", rr.Body.String())
	}
	assertScoreUnchanged(t, "Model1", 80)
}

func TestUpdateResultContract_MalformedFormRejected(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()
	seedUpdateResultFixture(t)

	req := httptest.NewRequest("POST", "/update_result", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	UpdateResultHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for malformed form body, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid form data") {
		t.Fatalf("expected invalid form data message, got %q", rr.Body.String())
	}
	assertScoreUnchanged(t, "Model1", 80)
}

func TestUpdateResultContract_NonNumericPromptIndexRejected(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()
	seedUpdateResultFixture(t)

	form := url.Values{}
	form.Add("model", "Model1")
	form.Add("promptIndex", "abc")
	form.Add("pass", "true")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	UpdateResultHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for non-numeric promptIndex, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid prompt index") {
		t.Fatalf("expected invalid prompt index message, got %q", rr.Body.String())
	}
	assertScoreUnchanged(t, "Model1", 80)
}

func TestUpdateResultContract_NegativePromptIndexRejected(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()
	seedUpdateResultFixture(t)

	form := url.Values{}
	form.Add("model", "Model1")
	form.Add("promptIndex", "-1")
	form.Add("pass", "true")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	UpdateResultHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for negative promptIndex, got %d", http.StatusBadRequest, rr.Code)
	}
	assertScoreUnchanged(t, "Model1", 80)
}

func TestUpdateResultContract_PromptIndexBeyondBoundsRejected(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()
	seedUpdateResultFixture(t)

	form := url.Values{}
	form.Add("model", "Model1")
	form.Add("promptIndex", "99")
	form.Add("pass", "true")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	UpdateResultHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for out-of-range promptIndex, got %d", http.StatusBadRequest, rr.Code)
	}
	assertScoreUnchanged(t, "Model1", 80)
}

func TestUpdateResultContract_MissingModelRejected(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()
	seedUpdateResultFixture(t)

	form := url.Values{}
	form.Add("promptIndex", "0")
	form.Add("pass", "true")

	req := httptest.NewRequest("POST", "/update_result", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	UpdateResultHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for missing model, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Missing model") {
		t.Fatalf("expected missing model message, got %q", rr.Body.String())
	}
	assertScoreUnchanged(t, "Model1", 80)
}
