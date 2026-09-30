package handlers

import (
	"fmt"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestTierAlgorithm exercises the production tier machinery directly: every
// score GetRandomScoreForTierWrapper emits must be a member of the tier value
// set {0, 20, 40, 60, 80, 100}, and the totals RandomizeScores writes for each
// model must land inside the band of the tier that model was assigned to.
func TestTierAlgorithm(t *testing.T) {
	validScores := map[int]bool{0: true, 20: true, 40: true, 60: true, 80: true, 100: true}

	for _, tier := range []int{-1, 0, 1, 5, 10, 11, 12, 100} {
		for i := 0; i < 100; i++ {
			if score := GetRandomScoreForTierWrapper(tier); !validScores[score] {
				t.Fatalf("GetRandomScoreForTierWrapper(%d) produced invalid score %d", tier, score)
			}
		}
	}

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()
	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		t.Fatalf("failed to get suite ID: %v", err)
	}

	numModels := 12
	numPrompts := 50
	numTiers := 12

	for i := 0; i < numPrompts; i++ {
		if _, err := db.Exec("INSERT INTO prompts (text, suite_id, display_order, type) VALUES (?, ?, ?, 'objective')",
			fmt.Sprintf("Prompt %d", i), suiteID, i); err != nil {
			t.Fatalf("failed to insert prompt %d: %v", i, err)
		}
	}
	for i := 0; i < numModels; i++ {
		if _, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)",
			fmt.Sprintf("Model%d", i), suiteID); err != nil {
			t.Fatalf("failed to insert model %d: %v", i, err)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/randomize_scores", nil)
	rr := httptest.NewRecorder()
	DefaultHandler.RandomizeScores(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	maxScore := numPrompts * 100

	// The handler walks models in row order, matching ascending ids here.
	rows, err := db.Query(`
		SELECT m.id, m.name, SUM(s.score) AS total
		FROM models m JOIN scores s ON s.model_id = m.id
		WHERE m.suite_id = ?
		GROUP BY m.id
		ORDER BY m.id`, suiteID)
	if err != nil {
		t.Fatalf("failed to query totals: %v", err)
	}
	defer func() { _ = rows.Close() }()

	type modelTotal struct {
		name  string
		total int
	}
	var totals []modelTotal
	for rows.Next() {
		var mt modelTotal
		var id int
		if err := rows.Scan(&id, &mt.name, &mt.total); err != nil {
			t.Fatalf("failed to scan total: %v", err)
		}
		totals = append(totals, mt)
	}
	if len(totals) != numModels {
		t.Fatalf("expected %d scored models, got %d", numModels, len(totals))
	}

	for i, mt := range totals {
		tierIndex := (i * numTiers) / len(totals)
		bandLow := maxScore * tierIndex / numTiers
		bandHigh := maxScore * (tierIndex + 1) / numTiers
		if mt.total < bandLow || mt.total > bandHigh {
			t.Errorf("model %s (tier %d): total %d outside band [%d, %d]",
				mt.name, tierIndex, mt.total, bandLow, bandHigh)
		}
	}

	// Every persisted score must be a member of the tier value set.
	scoreRows, err := db.Query(`
		SELECT s.score
		FROM scores s JOIN models m ON s.model_id = m.id
		WHERE m.suite_id = ?
		ORDER BY m.id`, suiteID)
	if err != nil {
		t.Fatalf("failed to query scores: %v", err)
	}
	defer func() { _ = scoreRows.Close() }()

	scoreCount := 0
	for scoreRows.Next() {
		var score int
		if err := scoreRows.Scan(&score); err != nil {
			t.Fatalf("failed to scan score: %v", err)
		}
		if !validScores[score] {
			t.Errorf("persisted score %d is not a member of {0, 20, 40, 60, 80, 100}", score)
		}
		scoreCount++
	}
	if scoreCount != numModels*numPrompts {
		t.Errorf("expected %d persisted scores, got %d", numModels*numPrompts, scoreCount)
	}
}
