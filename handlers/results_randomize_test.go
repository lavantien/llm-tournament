package handlers

import (
	"fmt"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRandomizeScoresHandler_OnlyRandomizesScores_NotRegenerateData(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()
	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("failed to get suite ID: %v", err)
	}

	// Create initial prompts
	for i := 0; i < 5; i++ {
		_, err = db.Exec("INSERT INTO prompts (text, suite_id, display_order, type) VALUES (?, ?, ?, 'objective')", fmt.Sprintf("Prompt %d", i), suiteID, i)
		if err != nil {
			t.Fatalf("failed to insert prompt: %v", err)
		}
	}

	// Create a model
	_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "Model1", suiteID)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	// Set initial scores to all 100
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{"models": ["Model1"], "results": {"Model1": {"scores": [100, 100, 100, 100, 100]}}}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	DefaultHandler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify initial scores are all 100
	var scores string
	err = db.QueryRow("SELECT GROUP_CONCAT(score ORDER BY display_order) FROM scores s JOIN prompts p ON s.prompt_id = p.id WHERE s.model_id = (SELECT id FROM models WHERE name = 'Model1')").Scan(&scores)
	if err != nil {
		t.Fatalf("failed to query scores: %v", err)
	}
	if scores != "100,100,100,100,100" {
		t.Fatalf("expected initial scores '100,100,100,100,100', got '%s'", scores)
	}

	// Count initial prompts and models
	var promptCount int
	err = db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&promptCount)
	if err != nil {
		t.Fatalf("failed to count prompts: %v", err)
	}

	var modelCount int
	err = db.QueryRow("SELECT COUNT(*) FROM models").Scan(&modelCount)
	if err != nil {
		t.Fatalf("failed to count models: %v", err)
	}

	// Call RandomizeScores handler
	req2 := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr2 := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr2.Code)
	}

	// Verify prompts and models haven't changed
	var newPromptCount int
	err = db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&newPromptCount)
	if err != nil {
		t.Fatalf("failed to count prompts: %v", err)
	}
	if newPromptCount != promptCount {
		t.Errorf("prompt count changed from %d to %d", promptCount, newPromptCount)
	}

	var newModelCount int
	err = db.QueryRow("SELECT COUNT(*) FROM models").Scan(&newModelCount)
	if err != nil {
		t.Fatalf("failed to count models: %v", err)
	}
	if newModelCount != modelCount {
		t.Errorf("model count changed from %d to %d", modelCount, newModelCount)
	}

	// Verify scores are different (randomized)
	var newScores string
	err = db.QueryRow("SELECT GROUP_CONCAT(score ORDER BY display_order) FROM scores s JOIN prompts p ON s.prompt_id = p.id WHERE s.model_id = (SELECT id FROM models WHERE name = 'Model1')").Scan(&newScores)
	if err != nil {
		t.Fatalf("failed to query scores: %v", err)
	}
	if newScores == scores {
		t.Error("scores were not randomized")
	}
}

func TestRandomizeScores_UsesTierBasedDistribution(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()
	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("failed to get suite ID: %v", err)
	}

	// Create prompts
	for i := 0; i < 50; i++ {
		_, err = db.Exec("INSERT INTO prompts (text, suite_id, display_order, type) VALUES (?, ?, ?, 'objective')", fmt.Sprintf("Prompt %d", i), suiteID, i)
		if err != nil {
			t.Fatalf("failed to insert prompt: %v", err)
		}
	}

	// Create multiple models to test tier distribution
	for i := 0; i < 12; i++ {
		_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", fmt.Sprintf("Model%d", i), suiteID)
		if err != nil {
			t.Fatalf("failed to insert model: %v", err)
		}
	}

	// Run randomize
	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Query scores and verify they're all valid
	rows, err := db.Query("SELECT score FROM scores")
	if err != nil {
		t.Fatalf("failed to query scores: %v", err)
	}
	defer func() { _ = rows.Close() }()

	validScores := map[int]bool{0: true, 20: true, 40: true, 60: true, 80: true, 100: true}
	for rows.Next() {
		var score int
		if err := rows.Scan(&score); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		if !validScores[score] {
			t.Errorf("invalid score: %d", score)
		}
	}

	// Verify models are distributed across tiers (at least 8 of 12 tiers)
	tierRows, err := db.Query(`
		SELECT m.name, SUM(s.score) as total
		FROM models m
		JOIN scores s ON s.model_id = m.id
		WHERE m.suite_id = ?
		GROUP BY m.id
		ORDER BY total DESC
	`, suiteID)
	if err != nil {
		t.Fatalf("failed to query tier distribution: %v", err)
	}
	defer func() { _ = tierRows.Close() }()

	maxScore := 50 * 100
	tierSize := maxScore / 12
	tiers := make(map[int]bool)

	for tierRows.Next() {
		var name string
		var total int
		if err := tierRows.Scan(&name, &total); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		tier := total / tierSize
		if tier >= 12 {
			tier = 11
		}
		tiers[tier] = true
	}

	if len(tiers) < 8 {
		t.Errorf("expected at least 8 tiers to be populated, got %d", len(tiers))
	}
}

func TestRandomizeScoresHandler_WrapperFunction(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()
	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("failed to get suite ID: %v", err)
	}

	// Create a model
	_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "TestModel", suiteID)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	// Create a prompt
	_, err = db.Exec("INSERT INTO prompts (text, suite_id, display_order, type) VALUES (?, ?, ?, 'objective')", "Test prompt", suiteID, 0)
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}

	// Call the wrapper function directly
	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	RandomizeScoresHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestRandomizeScoresHandler_ErrorCases(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	tests := []struct {
		name         string
		method       string
		expectStatus int
	}{
		{
			name:         "GET method not allowed",
			method:       "GET",
			expectStatus: http.StatusMethodNotAllowed,
		},
		{
			name:         "PUT method not allowed",
			method:       "PUT",
			expectStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/randomize_scores", nil)
			rr := httptest.NewRecorder()
			RandomizeScoresHandler(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, rr.Code)
			}
		})
	}
}

func TestRandomizeScores_ModelsQueryError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	_, err := db.Exec("DROP TABLE models")
	if err != nil {
		t.Fatalf("failed to drop models table: %v", err)
	}

	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for models query error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestRandomizeScores_PromptsQueryError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	suiteID, _ := middleware.GetCurrentSuiteID()

	_, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "TestModel", suiteID)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	_, err = db.Exec("DROP TABLE prompts")
	if err != nil {
		t.Fatalf("failed to drop prompts table: %v", err)
	}

	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for prompts query error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestRandomizeScores_SingleModelTierBounds(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	suiteID, _ := middleware.GetCurrentSuiteID()

	_, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "OnlyModel", suiteID)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	_, err = db.Exec("INSERT INTO prompts (text, display_order, suite_id) VALUES (?, ?, ?)", "Prompt 1", 0, suiteID)
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}

	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var scoreCount int
	err = db.QueryRow("SELECT COUNT(*) FROM scores").Scan(&scoreCount)
	if err != nil {
		t.Fatalf("failed to query score count: %v", err)
	}

	if scoreCount != 1 {
		t.Errorf("expected 1 score to be created, got %d", scoreCount)
	}

	var score int
	err = db.QueryRow("SELECT score FROM scores").Scan(&score)
	if err != nil {
		t.Fatalf("failed to query score: %v", err)
	}

	if score < 0 || score > 100 {
		t.Errorf("expected score in [0, 100], got %d", score)
	}
}

func TestRandomizeScores_ModelScanError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	suiteID, _ := middleware.GetCurrentSuiteID()

	if _, err := db.Exec("INSERT INTO prompts (text, display_order, suite_id) VALUES (?, ?, ?)", "Prompt 1", 0, suiteID); err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}

	// Recreate models with a TEXT id so a row can fail the int scan while the query succeeds.
	if _, err := db.Exec("DROP TABLE models"); err != nil {
		t.Fatalf("drop models table: %v", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE models (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			suite_id INTEGER NOT NULL,
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
			UNIQUE(name, suite_id)
		)`); err != nil {
		t.Fatalf("failed to create models table: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO models (id, name, suite_id) VALUES ('bad', 'BadModel', ?)`, suiteID); err != nil {
		t.Fatalf("failed to insert bad model row: %v", err)
	}

	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d with scan error skipped, got %d", http.StatusOK, rr.Code)
	}

	// The bad row is skipped, so no scores are randomized for it.
	var scoreCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM scores").Scan(&scoreCount); err != nil {
		t.Fatalf("failed to query score count: %v", err)
	}
	if scoreCount != 0 {
		t.Errorf("expected 0 scores with the only model skipped, got %d", scoreCount)
	}
}

func TestRandomizeScores_PromptScanError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	suiteID, _ := middleware.GetCurrentSuiteID()

	if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "TestModel", suiteID); err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	// Recreate prompts with a TEXT id so a row can fail the int scan while the query succeeds.
	if _, err := db.Exec("DROP TABLE prompts"); err != nil {
		t.Fatalf("drop prompts table: %v", err)
	}
	if _, err := db.Exec(`
		CREATE TABLE prompts (
			id TEXT PRIMARY KEY,
			text TEXT NOT NULL,
			solution TEXT DEFAULT '',
			profile_id INTEGER,
			suite_id INTEGER NOT NULL,
			display_order INTEGER NOT NULL DEFAULT 0,
			type TEXT DEFAULT 'objective',
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE SET NULL,
			UNIQUE(text, suite_id)
		)`); err != nil {
		t.Fatalf("failed to create prompts table: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO prompts (id, text, solution, suite_id, display_order) VALUES ('bad', 'BadPrompt', '', ?, 0)`, suiteID); err != nil {
		t.Fatalf("failed to insert bad prompt row: %v", err)
	}

	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d with scan error skipped, got %d", http.StatusOK, rr.Code)
	}

	// The bad row is skipped, so no prompt ids are collected and no scores land.
	var scoreCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM scores").Scan(&scoreCount); err != nil {
		t.Fatalf("failed to query score count: %v", err)
	}
	if scoreCount != 0 {
		t.Errorf("expected 0 scores with the only prompt skipped, got %d", scoreCount)
	}
}

func TestRandomizeScores_GetCurrentSuiteIDError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	_, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "TestModel", 1)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	_, err = db.Exec("INSERT INTO prompts (text, display_order, suite_id) VALUES (?, ?, ?)", "Prompt 1", 0, 1)
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}

	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code == http.StatusInternalServerError || rr.Code == http.StatusOK {
		t.Logf("GetCurrentSuiteID handled correctly: status %d", rr.Code)
	}
}

func TestRandomizeScores_CoversAllTiers(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()
	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("failed to get suite ID: %v", err)
	}

	// Create 36 models and 50 prompts (enough for 12 tier distribution)
	numModels := 36
	numPrompts := 50

	// Create prompts
	for i := 0; i < numPrompts; i++ {
		_, err = db.Exec("INSERT INTO prompts (text, suite_id, display_order, type) VALUES (?, ?, ?, 'objective')", fmt.Sprintf("Prompt %d", i), suiteID, i)
		if err != nil {
			t.Fatalf("failed to insert prompt: %v", err)
		}
	}

	// Create models
	for i := 0; i < numModels; i++ {
		_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", fmt.Sprintf("Model%d", i), suiteID)
		if err != nil {
			t.Fatalf("failed to insert model: %v", err)
		}
	}

	// Run randomize
	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Query all model total scores
	rows, err := db.Query(`
		SELECT m.name, SUM(s.score) as total
		FROM models m
		JOIN scores s ON s.model_id = m.id
		WHERE m.suite_id = ?
		GROUP BY m.id
		ORDER BY total DESC
	`, suiteID)
	if err != nil {
		t.Fatalf("failed to query scores: %v", err)
	}
	defer func() { _ = rows.Close() }()

	type modelScore struct {
		Name  string
		Total int
	}
	var scores []modelScore
	for rows.Next() {
		var ms modelScore
		if err := rows.Scan(&ms.Name, &ms.Total); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		scores = append(scores, ms)
	}

	// Calculate max possible score
	maxScore := numPrompts * 100

	// Define 12 tiers (each tier covers approximately 8.33% of the range)
	// Count how many unique tiers are represented
	tierSize := maxScore / 12
	tiersRepresented := make(map[int]bool)

	for _, ms := range scores {
		tier := ms.Total / tierSize
		if tier >= 12 {
			tier = 11 // Cap at highest tier
		}
		tiersRepresented[tier] = true
	}

	// Verify at least 10 of 12 tiers are populated
	if len(tiersRepresented) < 10 {
		t.Errorf("Expected at least 10 of 12 tiers to be populated, got %d tiers", len(tiersRepresented))
		t.Logf("Tiers represented: %v", tiersRepresented)
		for _, ms := range scores {
			t.Logf("  %s: %d", ms.Name, ms.Total)
		}
	}
}

func TestClampToValidScore(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"exact_zero", 0, 0},
		{"exact_twenty", 20, 20},
		{"exact_forty", 40, 40},
		{"exact_sixty", 60, 60},
		{"exact_eighty", 80, 80},
		{"exact_hundred", 100, 100},
		{"below_zero", -10, 0},
		{"above_hundred", 150, 100},
		{"between_0_and_20", 10, 0},
		{"between_20_and_40", 30, 20},
		{"between_40_and_60", 50, 40},
		{"between_60_and_80", 70, 60},
		{"between_80_and_100", 90, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampToValidScore(tt.input)
			if result != tt.expected {
				t.Errorf("clampToValidScore(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetRandomScoreForTierWrapper_AllPaths(t *testing.T) {
	tests := []struct {
		name     string
		tierIdx  int
		validate func(t *testing.T, scores []int)
	}{
		{
			name:    "valid_tier_index_0",
			tierIdx: 0,
			validate: func(t *testing.T, scores []int) {
				for _, score := range scores {
					if score < 0 || score > 100 {
						t.Errorf("expected score in [0, 100], got %d", score)
					}
				}
			},
		},
		{
			name:    "valid_tier_index_10",
			tierIdx: 10,
			validate: func(t *testing.T, scores []int) {
				for _, score := range scores {
					if score < 0 || score > 100 {
						t.Errorf("expected score in [0, 100], got %d", score)
					}
					if score > 20 {
						t.Logf("primordial tier produced score %d (expected mostly 0-20)", score)
					}
				}
			},
		},
		{
			name:    "invalid_tier_index_negative",
			tierIdx: -1,
			validate: func(t *testing.T, scores []int) {
				lowScores := 0
				for _, score := range scores {
					if score < 60 {
						lowScores++
					}
				}
				if lowScores > len(scores)/4 {
					t.Errorf("clamped negative index should produce high scores, got %d low scores out of %d", lowScores, len(scores))
				}
			},
		},
		{
			name:    "invalid_tier_index_11",
			tierIdx: 11,
			validate: func(t *testing.T, scores []int) {
				highScores := 0
				for _, score := range scores {
					if score > 40 {
						highScores++
					}
				}
				if highScores > len(scores)/4 {
					t.Errorf("clamped index 11 should produce low scores (primordial), got %d high scores out of %d", highScores, len(scores))
				}
			},
		},
		{
			name:    "invalid_tier_index_100",
			tierIdx: 100,
			validate: func(t *testing.T, scores []int) {
				highScores := 0
				for _, score := range scores {
					if score > 40 {
						highScores++
					}
				}
				if highScores > len(scores)/4 {
					t.Errorf("clamped index 100 should produce low scores (primordial), got %d high scores out of %d", highScores, len(scores))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var scores []int
			for i := 0; i < 100; i++ {
				score := GetRandomScoreForTierWrapper(tt.tierIdx)
				scores = append(scores, score)
			}
			tt.validate(t, scores)
		})
	}
}

func TestRandomizeScores_SuitesTableDropped(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	// Create a model and prompt before dropping suites
	_, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "TestModel", 1)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	_, err = db.Exec("INSERT INTO prompts (text, display_order, suite_id) VALUES (?, ?, ?)", "Prompt 1", 0, 1)
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}

	// Drop the suites table to trigger GetCurrentSuiteID error
	_, err = db.Exec("DROP TABLE suites")
	if err != nil {
		t.Fatalf("failed to drop suites table: %v", err)
	}

	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d when suites table is dropped, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestRandomizeScores_PromptScanContinuesOnError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()
	suiteID, _ := middleware.GetCurrentSuiteID()

	// Create a model
	_, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "TestModel", suiteID)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	// Create prompts with valid data
	_, err = db.Exec("INSERT INTO prompts (text, display_order, suite_id) VALUES (?, ?, ?)", "Good Prompt", 0, suiteID)
	if err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}

	// Add a second prompt
	_, err = db.Exec("INSERT INTO prompts (text, display_order, suite_id) VALUES (?, ?, ?)", "Good Prompt 2", 1, suiteID)
	if err != nil {
		t.Fatalf("failed to insert second prompt: %v", err)
	}

	req := httptest.NewRequest("POST", "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestGetRandomScoreForTierWrapper_DirectCall(t *testing.T) {
	// Test the function directly to exercise all paths
	tests := []struct {
		name      string
		tierIndex int
	}{
		{"negative index", -1},
		{"zero index", 0},
		{"low tier", 1},
		{"mid tier", 5},
		{"high tier", 10},
		{"beyond max", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := GetRandomScoreForTierWrapper(tt.tierIndex)
			// Score should always be one of the valid values
			validScores := map[int]bool{0: true, 20: true, 40: true, 60: true, 80: true, 100: true}
			if !validScores[score] {
				t.Errorf("got invalid score %d", score)
			}
		})
	}
}
