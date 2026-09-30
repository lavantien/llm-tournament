package handlers

import (
	"encoding/json"
	"llm-tournament/middleware"
	"llm-tournament/testutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// modelsFromRenderData extracts the Models ordering the results page passes to
// the template, which is the ranked order the scoreboard renders.
func modelsFromRenderData(t *testing.T, data interface{}) []string {
	t.Helper()
	field := reflect.ValueOf(data).FieldByName("Models")
	if !field.IsValid() || field.Kind() != reflect.Slice {
		t.Fatalf("expected Models slice on template data, got %#v", data)
	}
	models := make([]string, field.Len())
	for i := range models {
		models[i] = field.Index(i).String()
	}
	return models
}

// TestResultsHandler_RankingOrderStableAcrossRenders seeds several models with
// equal totals and requires the rendered ranking to stay identical (descending
// score, then alphabetical) across many handler invocations. Map iteration
// order feeds the sort, so only a name tie-breaker keeps this deterministic.
func TestResultsHandler_RankingOrderStableAcrossRenders(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}, {Text: "P2"}},
		Results: map[string]middleware.Result{
			"Top":   {Scores: []int{100, 100}},
			"Zeta":  {Scores: []int{60, 60}},
			"Alpha": {Scores: []int{60, 60}},
			"Mid":   {Scores: []int{80, 40}},
			"Beta":  {Scores: []int{40, 80}},
		},
	}

	want := []string{"Top", "Alpha", "Beta", "Mid", "Zeta"}
	for i := 0; i < 25; i++ {
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
		got := modelsFromRenderData(t, renderer.RenderCalls[0].Data)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("render %d: expected model order %v, got %v", i, want, got)
		}
	}
}

// TestUpdateMockResults_RankingOrderStableAcrossCalls requires the models
// array echoed by the mock-results endpoint to keep the same deterministic
// order (descending total, then alphabetical) across many calls.
func TestUpdateMockResults_RankingOrderStableAcrossCalls(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	handler := &Handler{
		DataStore: &MockDataStore{Prompts: []middleware.Prompt{{Text: "P1"}}},
		Renderer:  &MockRenderer{},
	}

	mockData := `{
		"results": {
			"Zeta":  {"scores": [60]},
			"Alpha": {"scores": [60]},
			"Mid":   {"scores": [60]},
			"Beta":  {"scores": [60]},
			"Top":   {"scores": [100]}
		},
		"models": ["Zeta", "Alpha", "Mid", "Beta", "Top"]
	}`

	want := []string{"Top", "Alpha", "Beta", "Mid", "Zeta"}
	for i := 0; i < 25; i++ {
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
		if !reflect.DeepEqual(resp.Models, want) {
			t.Fatalf("call %d: expected model order %v, got %v", i, want, resp.Models)
		}
	}
}
