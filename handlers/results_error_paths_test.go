package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"llm-tournament/middleware"
)

func TestUpdateMockResults_LastInsertIDError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	origLastInsertID := lastInsertID
	lastInsertID = func(sql.Result) (int64, error) { return 0, errors.New("last insert id failed") }
	t.Cleanup(func() { lastInsertID = origLastInsertID })

	req := httptest.NewRequest(http.MethodPost, "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// The suite row inserts, but the failed ID read skips its whole seeding
	// branch: no models end up in the second suite.
	var secondSuiteModels int
	if err := middleware.GetDB().QueryRow(`
		SELECT COUNT(*) FROM models WHERE suite_id = (SELECT id FROM suites WHERE name = 'Alternative Suite')`).Scan(&secondSuiteModels); err != nil {
		t.Fatalf("count second-suite models: %v", err)
	}
	if secondSuiteModels != 0 {
		t.Errorf("expected no second-suite models when its insert ID errors, got %d", secondSuiteModels)
	}
}

func TestUpdateMockResults_SecondSuiteProfileLastInsertIDError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Seeding call order: 5 first-suite profiles, 1 second-suite insert, then
	// 4 second-suite profiles. Let only the suite insert succeed so the
	// second-suite profile loop runs and hits its LastInsertId error branch.
	origLastInsertID := lastInsertID
	calls := 0
	lastInsertID = func(result sql.Result) (int64, error) {
		calls++
		if calls == 6 {
			return 2, nil
		}
		return 0, errors.New("last insert id failed")
	}
	t.Cleanup(func() { lastInsertID = origLastInsertID })

	req := httptest.NewRequest(http.MethodPost, "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// The suite insert succeeded (call 6) but its profile inserts failed,
	// so the second suite exists while the first suite was never seeded.
	var suites int
	if err := middleware.GetDB().QueryRow(
		"SELECT COUNT(*) FROM suites WHERE name = 'Alternative Suite'").Scan(&suites); err != nil {
		t.Fatalf("count suites: %v", err)
	}
	if suites != 1 {
		t.Errorf("expected the second suite to exist, got %d rows", suites)
	}
}

func TestUpdateMockResults_SecondSuitePromptScanError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	// Recreate prompts with a TEXT id so a row can fail the int scan while the query succeeds.
	// The suite FK is omitted so the row can reference suite 2 before the handler creates it.
	if _, err := db.Exec("DROP TABLE prompts"); err != nil {
		t.Fatalf("drop prompts: %v", err)
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
			UNIQUE(text, suite_id)
		)`); err != nil {
		t.Fatalf("create prompts: %v", err)
	}

	// Suite 1 is the default suite; the handler creates "Alternative Suite" as suite 2.
	if _, err := db.Exec(
		`INSERT INTO prompts (id, text, solution, suite_id, display_order) VALUES ('bad', 'scan-error-row', '', 2, -1)`); err != nil {
		t.Fatalf("insert bad prompt row: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// The doctored prompts table also makes the final WriteResults fail, so the
	// deterministic outcome is 500; the second-suite scan-error continue at the
	// SELECT id loop above is exercised on the way there.
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Error saving mock results") {
		t.Errorf("expected WriteResults failure body, got %q", rr.Body.String())
	}
}

func TestUpdateMockResults_SecondSuiteModelLookupError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Silently skip second-suite model inserts so the later lookups miss.
	if _, err := middleware.GetDB().Exec(`
		CREATE TRIGGER skip_second_suite_models
		BEFORE INSERT ON models
		WHEN NEW.suite_id = 2
		BEGIN
			SELECT RAISE(IGNORE);
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// The lookup misses skip score creation for every second-suite prompt.
	var secondSuiteScores int
	if err := middleware.GetDB().QueryRow(`
		SELECT COUNT(*) FROM scores
		WHERE prompt_id IN (SELECT id FROM prompts WHERE suite_id = 2)`).Scan(&secondSuiteScores); err != nil {
		t.Fatalf("count second-suite scores: %v", err)
	}
	if secondSuiteScores != 0 {
		t.Errorf("expected no scores for the second suite, got %d", secondSuiteScores)
	}
}

func TestUpdateMockResults_SecondSuiteScoreInsertError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Fail every score insert; both suites keep seeding and the score-insert
	// error branches (including the second suite's) log-and-continue. The
	// final WriteResults also inserts scores, so it deterministically fails.
	if _, err := middleware.GetDB().Exec(`
		CREATE TRIGGER fail_score_inserts
		BEFORE INSERT ON scores
		BEGIN
			SELECT RAISE(ABORT, 'score insert failed');
		END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Error saving mock results") {
		t.Errorf("expected WriteResults failure body, got %q", rr.Body.String())
	}

	var scoreCount int
	if err := middleware.GetDB().QueryRow("SELECT COUNT(*) FROM scores").Scan(&scoreCount); err != nil {
		t.Fatalf("count scores: %v", err)
	}
	if scoreCount != 0 {
		t.Errorf("expected no scores when every insert fails, got %d", scoreCount)
	}
}

func TestUpdateMockResults_PromptLookupContinuesWhenDBHasNoPrompts(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	if _, err := middleware.GetDB().Exec("INSERT INTO models (name, suite_id) VALUES ('TestModel', 1)"); err != nil {
		t.Fatalf("insert model: %v", err)
	}

	ds := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "p1"}, {Text: "p2"}, {Text: "p3"}},
		Results: map[string]middleware.Result{"TestModel": {Scores: []int{80, 60, 100}}},
	}
	handler := NewHandlerWithDeps(ds, &MockRenderer{})

	body := `{
		"results": {"TestModel": {"scores": [80, 60, 100]}},
		"models": ["TestModel"],
		"passPercentages": {},
		"totalScores": {}
	}`
	req := httptest.NewRequest(http.MethodPost, "/update_mock_results", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// Every prompt lookup missed, so no mock responses were generated.
	var responseCount int
	if err := middleware.GetDB().QueryRow("SELECT COUNT(*) FROM model_responses").Scan(&responseCount); err != nil {
		t.Fatalf("count model responses: %v", err)
	}
	if responseCount != 0 {
		t.Errorf("expected no model responses when prompt lookups fail, got %d", responseCount)
	}
}
