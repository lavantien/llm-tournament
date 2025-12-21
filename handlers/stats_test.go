package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"llm-tournament/middleware"
	"llm-tournament/testutil"

	_ "github.com/mattn/go-sqlite3"
)

// changeToProjectRootStats changes to the project root directory for tests that need templates
func changeToProjectRootStats(t *testing.T) func() {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}
	return func() {
		os.Chdir(originalDir)
	}
}

// setupStatsTestDB creates a test database for stats handler tests
func setupStatsTestDB(t *testing.T) func() {
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

func TestCalculateTiers_Empty(t *testing.T) {
	tiers, tierRanges := calculateTiers(map[string]int{})

	// Should return empty tiers
	for _, models := range tiers {
		if len(models) != 0 {
			t.Errorf("expected empty tier, got %v", models)
		}
	}

	// Tier ranges should always be populated
	if len(tierRanges) != 12 {
		t.Errorf("expected 12 tier ranges, got %d", len(tierRanges))
	}
}

func TestCalculateTiers_SingleModel(t *testing.T) {
	tests := []struct {
		name         string
		score        int
		expectedTier string
	}{
		{"transcendental minimum", 3780, "transcendental"},
		{"transcendental high", 5000, "transcendental"},
		{"cosmic maximum", 3779, "cosmic"},
		{"cosmic minimum", 3360, "cosmic"},
		{"divine maximum", 3359, "divine"},
		{"divine minimum", 2700, "divine"},
		{"celestial maximum", 2699, "celestial"},
		{"celestial minimum", 2400, "celestial"},
		{"ascendant maximum", 2399, "ascendant"},
		{"ascendant minimum", 2100, "ascendant"},
		{"ethereal maximum", 2099, "ethereal"},
		{"ethereal minimum", 1800, "ethereal"},
		{"mystic maximum", 1799, "mystic"},
		{"mystic minimum", 1500, "mystic"},
		{"astral maximum", 1499, "astral"},
		{"astral minimum", 1200, "astral"},
		{"spiritual maximum", 1199, "spiritual"},
		{"spiritual minimum", 900, "spiritual"},
		{"primal maximum", 899, "primal"},
		{"primal minimum", 600, "primal"},
		{"mortal maximum", 599, "mortal"},
		{"mortal minimum", 300, "mortal"},
		{"primordial maximum", 299, "primordial"},
		{"primordial minimum", 0, "primordial"},
		{"primordial negative", -100, "primordial"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tiers, _ := calculateTiers(map[string]int{"TestModel": tt.score})

			// Check the model is in the expected tier
			found := false
			for tier, models := range tiers {
				for _, model := range models {
					if model == "TestModel" {
						if tier != tt.expectedTier {
							t.Errorf("score %d: expected tier %q, got %q", tt.score, tt.expectedTier, tier)
						}
						found = true
					}
				}
			}
			if !found {
				t.Errorf("TestModel not found in any tier for score %d", tt.score)
			}
		})
	}
}

func TestCalculateTiers_AllBoundaries(t *testing.T) {
	// Test all boundary values
	boundaries := []struct {
		score int
		tier  string
	}{
		{3780, "transcendental"},
		{3779, "cosmic"},
		{3360, "cosmic"},
		{3359, "divine"},
		{2700, "divine"},
		{2699, "celestial"},
		{2400, "celestial"},
		{2399, "ascendant"},
		{2100, "ascendant"},
		{2099, "ethereal"},
		{1800, "ethereal"},
		{1799, "mystic"},
		{1500, "mystic"},
		{1499, "astral"},
		{1200, "astral"},
		{1199, "spiritual"},
		{900, "spiritual"},
		{899, "primal"},
		{600, "primal"},
		{599, "mortal"},
		{300, "mortal"},
		{299, "primordial"},
		{0, "primordial"},
	}

	for _, b := range boundaries {
		t.Run("", func(t *testing.T) {
			tiers, _ := calculateTiers(map[string]int{"Model": b.score})
			if len(tiers[b.tier]) != 1 || tiers[b.tier][0] != "Model" {
				t.Errorf("score %d should be in tier %q, but found in wrong tier", b.score, b.tier)
			}
		})
	}
}

func TestCalculateTiers_MultipleModels(t *testing.T) {
	totalScores := map[string]int{
		"GPT-4":   3800, // transcendental
		"Claude":  3500, // cosmic
		"Gemini":  2800, // divine
		"LLaMA":   1200, // astral
		"Mistral": 100,  // primordial
	}

	tiers, _ := calculateTiers(totalScores)

	// Verify each model is in the correct tier
	expectedTiers := map[string]string{
		"GPT-4":   "transcendental",
		"Claude":  "cosmic",
		"Gemini":  "divine",
		"LLaMA":   "astral",
		"Mistral": "primordial",
	}

	for model, expectedTier := range expectedTiers {
		found := false
		for _, m := range tiers[expectedTier] {
			if m == model {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("model %q should be in tier %q", model, expectedTier)
		}
	}
}

func TestCalculateTiers_TierRanges(t *testing.T) {
	_, tierRanges := calculateTiers(map[string]int{})

	expectedRanges := map[string]string{
		"transcendental": "3780+",
		"cosmic":         "3360-3779",
		"divine":         "2700-3359",
		"celestial":      "2400-2699",
		"ascendant":      "2100-2399",
		"ethereal":       "1800-2099",
		"mystic":         "1500-1799",
		"astral":         "1200-1499",
		"spiritual":      "900-1199",
		"primal":         "600-899",
		"mortal":         "300-599",
		"primordial":     "0-299",
	}

	for tier, expectedRange := range expectedRanges {
		if tierRanges[tier] != expectedRange {
			t.Errorf("tier %q: expected range %q, got %q", tier, expectedRange, tierRanges[tier])
		}
	}
}

func TestCalculateTiers_ModelsInSameTier(t *testing.T) {
	totalScores := map[string]int{
		"Model A": 3800,
		"Model B": 3900,
		"Model C": 4000,
	}

	tiers, _ := calculateTiers(totalScores)

	// All three models should be in transcendental
	if len(tiers["transcendental"]) != 3 {
		t.Errorf("expected 3 models in transcendental, got %d", len(tiers["transcendental"]))
	}

	// All other tiers should be empty
	for tier, models := range tiers {
		if tier != "transcendental" && len(models) != 0 {
			t.Errorf("expected tier %q to be empty, got %v", tier, models)
		}
	}
}

func TestCalculateTiers_ZeroScore(t *testing.T) {
	tiers, _ := calculateTiers(map[string]int{"ZeroModel": 0})

	if len(tiers["primordial"]) != 1 || tiers["primordial"][0] != "ZeroModel" {
		t.Error("zero score should be in primordial tier")
	}
}

func TestCalculateTiers_AllTiersCovered(t *testing.T) {
	// Create one model for each tier
	totalScores := map[string]int{
		"M_transcendental": 3780,
		"M_cosmic":         3360,
		"M_divine":         2700,
		"M_celestial":      2400,
		"M_ascendant":      2100,
		"M_ethereal":       1800,
		"M_mystic":         1500,
		"M_astral":         1200,
		"M_spiritual":      900,
		"M_primal":         600,
		"M_mortal":         300,
		"M_primordial":     0,
	}

	tiers, _ := calculateTiers(totalScores)

	// Each tier should have exactly one model
	expectedTiers := []string{
		"transcendental", "cosmic", "divine", "celestial",
		"ascendant", "ethereal", "mystic", "astral",
		"spiritual", "primal", "mortal", "primordial",
	}

	for _, tier := range expectedTiers {
		if len(tiers[tier]) != 1 {
			t.Errorf("expected 1 model in tier %q, got %d", tier, len(tiers[tier]))
		}
	}
}

func TestStatsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootStats(t)
	defer restoreDir()

	cleanup := setupStatsTestDB(t)
	defer cleanup()

	// Add test data
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt 1"},
		{Text: "Test Prompt 2"},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"TestModel": {Scores: []int{80, 100}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/stats", nil)
	rr := httptest.NewRecorder()
	StatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Statistics") {
		t.Error("expected 'Statistics' in response body")
	}
}

func TestStatsHandler_GET_WithScoreBreakdowns(t *testing.T) {
	restoreDir := changeToProjectRootStats(t)
	defer restoreDir()

	cleanup := setupStatsTestDB(t)
	defer cleanup()

	// Add test data with various scores to cover all score buckets
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
		{Text: "Prompt 3"},
		{Text: "Prompt 4"},
		{Text: "Prompt 5"},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"TestModel": {Scores: []int{20, 40, 60, 80, 100}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/stats", nil)
	rr := httptest.NewRecorder()
	StatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "TestModel") {
		t.Error("expected model name in response body")
	}
}

func TestStatsHandler_FixesScoreMismatchTotals(t *testing.T) {
	mockDS := &MockDataStore{
		Results: map[string]middleware.Result{
			// Include an invalid score so the summed total differs from the bucketed total.
			"ModelX": {Scores: []int{1, 20}},
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

	modelStats := totalScores.MapIndex(reflect.ValueOf("ModelX"))
	if !modelStats.IsValid() {
		t.Fatalf("expected ModelX in TotalScores")
	}

	total := modelStats.FieldByName("TotalScore")
	if !total.IsValid() {
		t.Fatalf("expected TotalScore field in ModelX stats")
	}
	if total.Int() != 20 {
		t.Fatalf("expected TotalScore to be corrected to 20, got %d", total.Int())
	}
}

func TestStatsHandler_GET_EmptyResults(t *testing.T) {
	restoreDir := changeToProjectRootStats(t)
	defer restoreDir()

	cleanup := setupStatsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/stats", nil)
	rr := httptest.NewRecorder()
	StatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestStatsHandler_GET_MultipleModels(t *testing.T) {
	restoreDir := changeToProjectRootStats(t)
	defer restoreDir()

	cleanup := setupStatsTestDB(t)
	defer cleanup()

	// Add test data with multiple models
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Prompt 1"},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"GPT-4":  {Scores: []int{100}},
		"Claude": {Scores: []int{80}},
		"Gemini": {Scores: []int{60}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/stats", nil)
	rr := httptest.NewRecorder()
	StatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "GPT-4") {
		t.Error("expected GPT-4 in response body")
	}
}

func TestStatsHandler_GET_RenderError(t *testing.T) {
	cleanup := setupStatsTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/stats", nil)
	rr := httptest.NewRecorder()
	StatsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}
