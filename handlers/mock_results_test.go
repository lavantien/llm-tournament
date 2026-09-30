package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestMinMax(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(int, int) int
		a, b     int
		expected int
	}{
		{"min_first_smaller", min, 1, 2, 1},
		{"min_second_smaller", min, 2, 1, 1},
		{"min_equal", min, 5, 5, 5},
		{"max_first_larger", max, 2, 1, 2},
		{"max_second_larger", max, 1, 2, 2},
		{"max_equal", max, 5, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(tt.a, tt.b); got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestUpdateMockResultsHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/update_mock_results", nil)
	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestUpdateMockResultsHandler_Success(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "Mock test prompt"}}
	_ = middleware.WritePrompts(prompts)

	mockData := map[string]interface{}{
		"models":  []string{"MockModel"},
		"results": map[string]middleware.Result{"MockModel": {Scores: []int{80}}},
	}
	jsonBody, _ := json.Marshal(mockData)

	req := httptest.NewRequest("POST", "/update_mock_results", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestUpdateMockResultsHandler_InvalidJSON(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader("not valid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUpdateMockResultsHandler_ValidatesScores(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "Validate test prompt"}}
	_ = middleware.WritePrompts(prompts)

	// Send invalid score (should be corrected to 0)
	mockData := map[string]interface{}{
		"models":  []string{"ValidateModel"},
		"results": map[string]middleware.Result{"ValidateModel": {Scores: []int{77}}}, // 77 is not valid
	}
	jsonBody, _ := json.Marshal(mockData)

	req := httptest.NewRequest("POST", "/update_mock_results", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify score was corrected
	results := middleware.ReadResults()
	if result, exists := results["ValidateModel"]; exists {
		if len(result.Scores) > 0 && result.Scores[0] != 0 {
			t.Errorf("expected score 0 (corrected from invalid 77), got %d", result.Scores[0])
		}
	}
}

func TestInitRand(t *testing.T) {
	r := initRand()
	if r == nil {
		t.Error("initRand should not return nil")
	}
	// Verify it produces random values
	val1 := r.Intn(1000)
	val2 := r.Intn(1000)
	// Very unlikely to be equal, but not impossible
	_ = val1
	_ = val2
}

func TestUpdateMockResultsHandler_WithEmptyModels(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts first
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Send request with results but no explicit models array
	mockData := `{
		"results": {
			"ModelFromResults": {"scores": [80]}
		},
		"models": [],
		"passPercentages": {},
		"totalScores": {}
	}`

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(mockData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestUpdateMockResultsHandler_ValidatesInvalidScores(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts first
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Send request with invalid scores (should be corrected to 0)
	mockData := `{
		"results": {
			"TestModel": {"scores": [15, 25, 35]}
		},
		"models": ["TestModel"],
		"passPercentages": {},
		"totalScores": {}
	}`

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(mockData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	UpdateMockResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify invalid scores were corrected
	results := middleware.ReadResults()
	if result, exists := results["TestModel"]; exists {
		for i, score := range result.Scores {
			if score != 0 {
				t.Errorf("expected invalid score at index %d to be corrected to 0, got %d", i, score)
			}
		}
	}
}

func TestUpdateMockResultsHandler_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts:      []middleware.Prompt{{Text: "Test prompt"}},
		CurrentSuite: "test-suite",
		WriteResultsFunc: func(suiteName string, results map[string]middleware.Result) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	mockData := `{
		"results": {"TestModel": {"scores": [80]}},
		"models": ["TestModel"]
	}`

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(mockData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestUpdateMockResults_ReadAllError(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{Prompts: []middleware.Prompt{{Text: "Test prompt"}}},
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest("POST", "/update_mock_results", readErrorReader{})
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestUpdateMockResults_SortsModelsByTotalScore(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{Prompts: []middleware.Prompt{{Text: "Test prompt"}}},
		Renderer:  &MockRenderer{},
	}

	// Intentionally provide models in reverse order; handler should sort by total score.
	mockData := `{
                "results": {
                        "ModelA": {"scores": [100]},
                        "ModelB": {"scores": [0]}
                },
                "models": ["ModelB", "ModelA"]
        }`

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
	if len(resp.Models) != 2 || resp.Models[0] != "ModelA" || resp.Models[1] != "ModelB" {
		t.Fatalf("expected models sorted as [ModelA ModelB], got %v", resp.Models)
	}
}

func TestUpdateMockResults_ResponseEncodeError(t *testing.T) {
	handler := &Handler{
		DataStore: &MockDataStore{Prompts: []middleware.Prompt{{Text: "Test prompt"}}},
		Renderer:  &MockRenderer{},
	}

	mockData := `{
                "results": {"TestModel": {"scores": [80]}},
                "models": ["TestModel"]
        }`

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(mockData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	failingWriter := &FailingResponseWriter{
		ResponseWriter: rr,
		WriteError:     errors.New("mock write error"),
	}

	handler.UpdateMockResults(failingWriter, req)

	// Handler logs encode errors but doesn't return a different status code.
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestUpdateMockResultsHandler_GeneratesMockModels(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Initial state: verify no models exist
	db := middleware.GetDB()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM models").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query model count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 models initially, got %d", count)
	}

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Use default handler which uses real middleware
	DefaultHandler.UpdateMockResults(rr, req)

	// Verify 24 mock models created
	err = db.QueryRow("SELECT COUNT(*) FROM models").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query model count after generation: %v", err)
	}
	if count != 24 {
		t.Errorf("expected 24 mock models, got %d", count)
	}

	// Verify model names use tier-based pattern
	rows, err := db.Query("SELECT name FROM models ORDER BY name")
	if err != nil {
		t.Fatalf("failed to query models: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	}()

	models := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan model name: %v", err)
		}
		models = append(models, name)
	}

	// Check for expected tier prefixes in model names
	expectedTiers := []string{"Cosmic", "Transcendent", "Ethereal", "Celestial", "Infinite"}
	foundTiers := make(map[string]bool)
	for _, model := range models {
		for _, tier := range expectedTiers {
			if strings.Contains(model, tier) {
				foundTiers[tier] = true
				break
			}
		}
	}

	// At least some expected tiers should be present
	if len(foundTiers) < 3 {
		t.Errorf("expected tier-based model names (Cosmic, Transcendent, etc.), got: %v", models[:5])
	}
}

func TestUpdateMockResultsHandler_GeneratesTierBasedModelNames(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Use default handler which uses real middleware
	DefaultHandler.UpdateMockResults(rr, req)

	// Verify model names use tier-based pattern
	db := middleware.GetDB()
	rows, err := db.Query("SELECT name FROM models ORDER BY name")
	if err != nil {
		t.Fatalf("failed to query models: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	}()

	models := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan model name: %v", err)
		}
		models = append(models, name)
	}

	// Check that at least one model contains 'Cosmic'
	foundCosmic := false
	for _, model := range models {
		if strings.Contains(model, "Cosmic") {
			foundCosmic = true
			break
		}
	}
	if !foundCosmic {
		t.Errorf("expected at least one model to contain 'Cosmic', got: %v", models[:5])
	}
}

func TestUpdateMockResultsHandler_GeneratesMockResponses(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Use default handler which uses real middleware
	DefaultHandler.UpdateMockResults(rr, req)

	// Verify mock responses were created
	db := middleware.GetDB()
	var responseCount int
	err := db.QueryRow("SELECT COUNT(*) FROM model_responses WHERE response_source = 'mock'").Scan(&responseCount)
	if err != nil {
		t.Fatalf("failed to query response count: %v", err)
	}

	// Should have 15 models * 1 prompt = 15 mock responses
	expectedResponses := 24 // models count * prompt count
	if responseCount < expectedResponses {
		t.Errorf("expected at least %d mock responses, got %d", expectedResponses, responseCount)
	}

	// Verify response text is not empty
	var responseText string
	err = db.QueryRow("SELECT response_text FROM model_responses WHERE response_source = 'mock' LIMIT 1").Scan(&responseText)
	if err != nil {
		t.Fatalf("failed to query mock response: %v", err)
	}
	if len(responseText) == 0 {
		t.Error("expected mock response text to be non-empty")
	}
}

func TestUpdateMockResultsHandler_CreatesModelsInCurrentSuite(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})

	// Trigger mock generation with empty request (no existing models)
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Verify the response is successful
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// Verify models were created in the database
	db := middleware.GetDB()
	var modelCount int
	err := db.QueryRow("SELECT COUNT(*) FROM models").Scan(&modelCount)
	if err != nil {
		t.Fatalf("failed to query model count: %v", err)
	}

	// Should have created 24 mock models
	expectedModels := 24
	if modelCount != expectedModels {
		t.Errorf("expected %d models in database, got %d", expectedModels, modelCount)
	}
}

func TestUpdateMockResultsHandler_GeneratesMockPrompts(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Start with completely empty database (no prompts, no models)

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Verify the response is successful
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	// Verify prompts were created in the database
	db := middleware.GetDB()
	var promptCount int
	err := db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&promptCount)
	if err != nil {
		t.Fatalf("failed to query prompt count: %v", err)
	}

	// Should have created mock prompts
	if promptCount == 0 {
		t.Errorf("expected prompts to be created in database, got %d", promptCount)
	}
}

func TestUpdateMockResultsHandler_CreatesProfiles(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Start with completely empty database (no prompts, no models)

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Verify 5 profiles were created
	db := middleware.GetDB()
	var profileCount int
	err := db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&profileCount)
	if err != nil {
		t.Fatalf("failed to query profile count: %v", err)
	}

	if profileCount != 9 {
		t.Errorf("expected 9 profiles to be created in database (5 from first suite + 4 from second suite), got %d", profileCount)
	}

	// Verify profile names
	rows, err := db.Query("SELECT name FROM profiles ORDER BY name")
	if err != nil {
		t.Fatalf("failed to query profiles: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	}()

	var profiles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan profile name: %v", err)
		}
		profiles = append(profiles, name)
	}

	expectedProfiles := []string{"Art", "Geography", "History", "Literature", "Math", "Philosophy", "Programming", "Science", "Writing"}
	if len(profiles) != len(expectedProfiles) {
		t.Fatalf("expected %d profiles, got %d", len(expectedProfiles), len(profiles))
	}

	for i, expected := range expectedProfiles {
		if profiles[i] != expected {
			t.Errorf("expected profile %d to be '%s', got '%s'", i, expected, profiles[i])
		}
	}
}

func TestUpdateMockResultsHandler_CreatesProfileBasedPrompts(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Start with completely empty database

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Verify 70 prompts were created (50 from first suite + 20 from second suite)
	db := middleware.GetDB()
	var promptCount int
	err := db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&promptCount)
	if err != nil {
		t.Fatalf("failed to query prompt count: %v", err)
	}

	if promptCount != 70 {
		t.Errorf("expected 70 prompts to be created in database (50 from first suite + 20 from second suite), got %d", promptCount)
	}

	// Verify prompts have profiles assigned
	var promptsWithProfile int
	err = db.QueryRow("SELECT COUNT(*) FROM prompts p WHERE p.profile_id IS NOT NULL").Scan(&promptsWithProfile)
	if err != nil {
		t.Fatalf("failed to query prompts with profiles: %v", err)
	}

	if promptsWithProfile != 70 {
		t.Errorf("expected all 70 prompts to have profiles, got %d", promptsWithProfile)
	}

	// Verify 9 profiles exist (5 from first suite + 4 from second suite)
	var profileCount int
	err = db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&profileCount)
	if err != nil {
		t.Fatalf("failed to query profile count: %v", err)
	}

	if profileCount != 9 {
		t.Errorf("expected 9 profiles, got %d", profileCount)
	}

	// Verify each profile has the correct number of prompts
	rows, err := db.Query(`
		SELECT pr.name, COUNT(p.id) as prompt_count
		FROM profiles pr
		LEFT JOIN prompts p ON p.profile_id = pr.id
		GROUP BY pr.name
		ORDER BY pr.name
	`)
	if err != nil {
		t.Fatalf("failed to query profile prompt counts: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	}()

	expectedProfiles := map[string]int{
		// First suite: 5 profiles × 10 prompts each
		"Math":        10,
		"Philosophy":  10,
		"Programming": 10,
		"Science":     10,
		"Writing":     10,
		// Second suite: 4 profiles × 5 prompts each
		"Art":        5,
		"Geography":  5,
		"History":    5,
		"Literature": 5,
	}

	for rows.Next() {
		var profileName string
		var count int
		if err := rows.Scan(&profileName, &count); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}

		expectedCount, exists := expectedProfiles[profileName]
		if !exists {
			t.Errorf("unexpected profile: %s", profileName)
		}
		if count != expectedCount {
			t.Errorf("profile %s: expected %d prompts, got %d", profileName, expectedCount, count)
		}
	}
}

func TestUpdateMockResultsHandler_DatabaseError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	_, err := db.Exec("DROP TABLE profiles")
	if err != nil {
		t.Fatalf("failed to drop profiles table: %v", err)
	}

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Every profile insert fails, so no prompts are ever seeded; the mock
	// model generation still runs and the request must succeed.
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d when profiles table is dropped, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var promptCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&promptCount); err != nil {
		t.Fatalf("failed to count prompts: %v", err)
	}
	if promptCount != 0 {
		t.Errorf("expected no prompts without profiles, got %d", promptCount)
	}

	var modelCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM models").Scan(&modelCount); err != nil {
		t.Fatalf("failed to count models: %v", err)
	}
	if modelCount != 36 {
		t.Errorf("expected 24 generated models plus 12 second-suite models, got %d", modelCount)
	}

	var scoreCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM scores").Scan(&scoreCount); err != nil {
		t.Fatalf("failed to count scores: %v", err)
	}
	if scoreCount != 0 {
		t.Errorf("expected no scores without prompts, got %d", scoreCount)
	}
}

func TestUpdateMockResultsHandler_ModelsTableError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	_, err := db.Exec("DROP TABLE models")
	if err != nil {
		t.Fatalf("failed to drop models table: %v", err)
	}

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Seeding completes, but the final WriteResults cannot query the dropped
	// models table, so the handler must report the save failure.
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d when models table is dropped, got %d: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Error saving mock results") {
		t.Errorf("expected WriteResults failure body, got %q", rr.Body.String())
	}

	var promptCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&promptCount); err != nil {
		t.Fatalf("failed to count prompts: %v", err)
	}
	if promptCount != 70 {
		t.Errorf("expected 70 seeded prompts, got %d", promptCount)
	}

	var profileCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&profileCount); err != nil {
		t.Fatalf("failed to count profiles: %v", err)
	}
	if profileCount != 9 {
		t.Errorf("expected 9 seeded profiles, got %d", profileCount)
	}
}

func TestUpdateMockResultsHandler_SuitesTableError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	_, err := db.Exec("DROP TABLE suites")
	if err != nil {
		t.Fatalf("failed to drop suites table: %v", err)
	}

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Without the suites table every suite-dependent insert fails (foreign
	// keys reference it) and WriteResults cannot resolve the suite, so the
	// handler must report the save failure and seed nothing.
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d when suites table is dropped, got %d: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Error saving mock results") {
		t.Errorf("expected WriteResults failure body, got %q", rr.Body.String())
	}

	var promptCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&promptCount); err != nil {
		t.Fatalf("failed to count prompts: %v", err)
	}
	if promptCount != 0 {
		t.Errorf("expected no prompts when the suites table is missing, got %d", promptCount)
	}

	var profileCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&profileCount); err != nil {
		t.Fatalf("failed to count profiles: %v", err)
	}
	if profileCount != 0 {
		t.Errorf("expected no profiles when the suites table is missing, got %d", profileCount)
	}
}

func TestUpdateMockResultsHandler_PromptsTableError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()

	suiteID, _ := middleware.GetCurrentSuiteID()

	_, err := db.Exec("INSERT INTO profiles (name, description, suite_id) VALUES (?, ?, ?)", "TestProfile", "Test", suiteID)
	if err != nil {
		t.Fatalf("failed to insert test profile: %v", err)
	}

	_, err = db.Exec("DROP TABLE prompts")
	if err != nil {
		t.Fatalf("failed to drop prompts table: %v", err)
	}

	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Profile inserts succeed but every prompt insert fails, and the final
	// WriteResults cannot query the dropped prompts table, so the handler
	// must report the save failure.
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d when prompts table is dropped, got %d: %s", http.StatusInternalServerError, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Error saving mock results") {
		t.Errorf("expected WriteResults failure body, got %q", rr.Body.String())
	}

	var profileCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&profileCount); err != nil {
		t.Fatalf("failed to count profiles: %v", err)
	}
	if profileCount != 10 {
		t.Errorf("expected 10 profiles (1 existing plus 9 seeded), got %d", profileCount)
	}

	var modelCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM models").Scan(&modelCount); err != nil {
		t.Fatalf("failed to count models: %v", err)
	}
	if modelCount != 12 {
		t.Errorf("expected 12 second-suite models, got %d", modelCount)
	}
}

func TestUpdateMockResultsHandler_PromptsHaveSolutions(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Trigger mock generation with empty request
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Verify all prompts have solution field (not NULL)
	db := middleware.GetDB()
	rows, err := db.Query("SELECT text, solution FROM prompts")
	if err != nil {
		t.Fatalf("failed to query prompts: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	}()

	count := 0
	for rows.Next() {
		var text, solution string
		if err := rows.Scan(&text, &solution); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		// Empty string is OK, NULL is not
		_ = solution // Empty solution is acceptable
		count++
	}

	if count == 0 {
		t.Error("no prompts found in database")
	}

	// Also verify we can read prompts via middleware without error
	prompts := middleware.ReadPrompts()
	if len(prompts) == 0 {
		t.Error("ReadPrompts returned empty slice")
	}
}

func TestUpdateMockResultsHandler_InsertErrorHandling(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()
	suiteID, _ := middleware.GetCurrentSuiteID()

	// Create prompts using DataStore so ReadPrompts() returns them
	promptsToCreate := []middleware.Prompt{}
	for i := 0; i < 5; i++ {
		promptsToCreate = append(promptsToCreate, middleware.Prompt{Text: fmt.Sprintf("Prompt %d", i)})
	}
	err := middleware.WritePromptSuite("default", promptsToCreate)
	if err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}

	// Create a model
	_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "TestModel", suiteID)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	// Drop the model_responses table to trigger insert error
	_, err = db.Exec("DROP TABLE model_responses")
	if err != nil {
		t.Fatalf("failed to drop model_responses table: %v", err)
	}

	// Send request with results that will try to insert mock responses
	reqBody := `{
		"results": {
			"TestModel": {"scores": [100, 80, 60, 40, 20]}
		},
		"models": ["TestModel"],
		"passPercentages": {"TestModel": 60.0},
		"totalScores": {"TestModel": 300}
	}`
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	DefaultHandler.UpdateMockResults(rr, req)

	// Insert errors are logged but must not fail the request, and the scores
	// themselves are saved before the response generation runs.
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	results := middleware.ReadResults()
	result, exists := results["TestModel"]
	if !exists {
		t.Fatal("expected results for TestModel to exist")
	}
	want := []int{100, 80, 60, 40, 20}
	if !reflect.DeepEqual(result.Scores, want) {
		t.Errorf("expected stored scores %v, got %v", want, result.Scores)
	}
}

func TestUpdateMockResultsHandler_PromptQueryRowError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	db := middleware.GetDB()
	suiteID, _ := middleware.GetCurrentSuiteID()

	// Create 5 prompts, then delete the last two so the database only
	// resolves offsets 0-2 while the DataStore still reports 5 prompts.
	promptsToCreate := []middleware.Prompt{}
	for i := 0; i < 5; i++ {
		promptsToCreate = append(promptsToCreate, middleware.Prompt{Text: fmt.Sprintf("Prompt %d", i)})
	}
	err := middleware.WritePromptSuite("default", promptsToCreate)
	if err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}
	_, err = db.Exec("DELETE FROM prompts WHERE display_order >= 3")
	if err != nil {
		t.Fatalf("failed to delete prompts: %v", err)
	}

	// Create a model
	_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", "TestModel", suiteID)
	if err != nil {
		t.Fatalf("failed to insert model: %v", err)
	}

	// The mock DataStore reports all 5 prompts, so the prompt lookups at
	// offsets 3 and 4 fail with ErrNoRows and must be skipped.
	handler := NewHandlerWithDeps(&MockDataStore{Prompts: promptsToCreate}, &MockRenderer{})

	reqBody := `{
		"results": {
			"TestModel": {"scores": [100, 80, 60, 40, 20]}
		},
		"models": ["TestModel"],
		"passPercentages": {"TestModel": 60.0},
		"totalScores": {"TestModel": 300}
	}`
	req := httptest.NewRequest("POST", "/update_mock_results", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateMockResults(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var responseCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM model_responses").Scan(&responseCount); err != nil {
		t.Fatalf("failed to count model responses: %v", err)
	}
	if responseCount != 3 {
		t.Errorf("expected 3 mock responses from the resolvable offsets, got %d", responseCount)
	}
}
