package handlers

import (
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
		preview := models
		if len(preview) > 5 {
			preview = preview[:5]
		}
		t.Errorf("expected tier-based model names (Cosmic, Transcendent, etc.), got: %v", preview)
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
		preview := models
		if len(preview) > 5 {
			preview = preview[:5]
		}
		t.Errorf("expected at least one model to contain 'Cosmic', got: %v", preview)
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
