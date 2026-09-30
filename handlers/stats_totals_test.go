package handlers

import (
	"llm-tournament/middleware"
	"llm-tournament/testutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// statsForModel extracts the ScoreStats entry rendered for one model
func statsForModel(t *testing.T, renderer *testutil.MockRenderer, model string) reflect.Value {
	t.Helper()
	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}
	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Struct {
		t.Fatalf("expected struct template data, got %T", data)
	}
	totalScores := val.FieldByName("TotalScores")
	if !totalScores.IsValid() || totalScores.Kind() != reflect.Map {
		t.Fatalf("expected TotalScores map on template data")
	}
	modelStats := totalScores.MapIndex(reflect.ValueOf(model))
	if !modelStats.IsValid() {
		t.Fatalf("expected %s in TotalScores", model)
	}
	return modelStats
}

func TestStatsHandler_TotalScoreIncludesNonStandardScores(t *testing.T) {
	cleanup := setupStatsTestDB(t)
	defer cleanup()

	if err := middleware.WritePromptSuite("default", []middleware.Prompt{{Text: "Prompt 1"}}); err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	// The evaluate page saves arbitrary 0-100 scores; 73 must count toward
	// the total even though it lands in no bucket.
	mockDS := &MockDataStore{
		Results: map[string]middleware.Result{
			"ModelX": {Scores: []int{73, 20}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	req := httptest.NewRequest("GET", "/stats", nil)
	rr := httptest.NewRecorder()
	handler.Stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	modelStats := statsForModel(t, renderer, "ModelX")
	if got := modelStats.FieldByName("TotalScore").Int(); got != 93 {
		t.Errorf("expected TotalScore 93 counting the 73 score, got %d", got)
	}
}

func TestStatsHandler_BucketsTrackOnlyStandardValues(t *testing.T) {
	cleanup := setupStatsTestDB(t)
	defer cleanup()

	if err := middleware.WritePromptSuite("default", []middleware.Prompt{{Text: "Prompt 1"}}); err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	mockDS := &MockDataStore{
		Results: map[string]middleware.Result{
			"ModelX": {Scores: []int{73, 20, 100}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	req := httptest.NewRequest("GET", "/stats", nil)
	rr := httptest.NewRecorder()
	handler.Stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	modelStats := statsForModel(t, renderer, "ModelX")
	expected := map[string]int64{
		"Count20":  1,
		"Count40":  0,
		"Count60":  0,
		"Count80":  0,
		"Count100": 1,
	}
	for field, want := range expected {
		if got := modelStats.FieldByName(field).Int(); got != want {
			t.Errorf("expected %s %d, got %d", field, want, got)
		}
	}
	if got := modelStats.FieldByName("TotalScore").Int(); got != 193 {
		t.Errorf("expected TotalScore 193 counting the 73 score, got %d", got)
	}
}
