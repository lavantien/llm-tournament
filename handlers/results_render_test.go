package handlers

import (
	"errors"
	"llm-tournament/middleware"
	"llm-tournament/testutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestResultsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test prompts
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt 1"},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	// Add a model
	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"TestModel": {Scores: []int{80}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	ResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "TestModel") {
		t.Error("expected model name in response body")
	}
}

func TestResultsHandler_GET_WithUncategorizedPrompts(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add prompts with empty profile (uncategorized)
	err := middleware.WritePrompts([]middleware.Prompt{
		{Text: "Categorized Prompt", Profile: "TestProfile"},
		{Text: "Uncategorized Prompt", Profile: ""},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	// Add a profile
	err = middleware.WriteProfiles([]middleware.Profile{
		{Name: "TestProfile", Description: "Test"},
	})
	if err != nil {
		t.Fatalf("failed to write test profiles: %v", err)
	}

	// Add results
	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"TestModel": {Scores: []int{80, 60}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	ResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Uncategorized") {
		t.Error("expected 'Uncategorized' group in response body for prompts without profile")
	}
}

func TestResultsHandler_GET_WithMultipleProfiles(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add multiple profiles
	err := middleware.WriteProfiles([]middleware.Profile{
		{Name: "Profile1", Description: "First profile"},
		{Name: "Profile2", Description: "Second profile"},
	})
	if err != nil {
		t.Fatalf("failed to write test profiles: %v", err)
	}

	// Add prompts across profiles
	err = middleware.WritePrompts([]middleware.Prompt{
		{Text: "Prompt 1", Profile: "Profile1"},
		{Text: "Prompt 2", Profile: "Profile1"},
		{Text: "Prompt 3", Profile: "Profile2"},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	// Add results
	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"Model1": {Scores: []int{100, 80, 60}},
		"Model2": {Scores: []int{60, 40, 20}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	ResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Profile1") {
		t.Error("expected Profile1 in response body")
	}
	if !strings.Contains(body, "Profile2") {
		t.Error("expected Profile2 in response body")
	}
}

func TestResultsHandler_ProfileGroups_UncategorizedStartColAfterProfile(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Categorized Prompt", Profile: "Profile1"},
			{Text: "Uncategorized Prompt", Profile: ""},
		},
		Profiles: []middleware.Profile{
			{Name: "Profile1", Description: "First profile"},
		},
		Results: map[string]middleware.Result{
			"Model1": {Scores: []int{100, 80}},
		},
	}
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

	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Struct {
		t.Fatalf("expected struct template data, got %T", data)
	}

	field := val.FieldByName("ProfileGroups")
	if !field.IsValid() {
		t.Fatalf("expected ProfileGroups field on template data")
	}

	var uncategorized *middleware.ProfileGroup
	for i := 0; i < field.Len(); i++ {
		pg, ok := field.Index(i).Interface().(*middleware.ProfileGroup)
		if !ok {
			t.Fatalf("expected *middleware.ProfileGroup, got %T", field.Index(i).Interface())
		}
		if pg.Name == "Uncategorized" {
			uncategorized = pg
			break
		}
	}

	if uncategorized == nil {
		t.Fatalf("expected Uncategorized profile group")
	}
	if uncategorized.StartCol != 1 {
		t.Fatalf("expected Uncategorized StartCol=1, got %d", uncategorized.StartCol)
	}
}

func TestResultsHandler_ModelFilterFiltersResultsMap(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
		},
		Results: map[string]middleware.Result{
			"Model1": {Scores: []int{100}},
			"Model2": {Scores: []int{80}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	req := httptest.NewRequest("GET", "/results?model_filter=Model1", nil)
	rr := httptest.NewRecorder()
	handler.Results(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}

	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)

	resultsField := val.FieldByName("Results")
	if resultsField.Kind() != reflect.Map {
		t.Fatalf("expected Results to be a map, got %v", resultsField.Kind())
	}
	if resultsField.Len() != 1 {
		t.Fatalf("expected 1 filtered result, got %d", resultsField.Len())
	}
	if !resultsField.MapIndex(reflect.ValueOf("Model1")).IsValid() {
		t.Fatalf("expected Model1 to remain after filtering")
	}
}

func TestResultsHandler_SearchQueryFiltersResultsMap(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
		},
		Results: map[string]middleware.Result{
			"AlphaModel": {Scores: []int{100}},
			"Beta":       {Scores: []int{80}},
		},
	}
	renderer := &testutil.MockRenderer{}
	handler := NewHandlerWithDeps(mockDS, renderer)

	req := httptest.NewRequest("GET", "/results?search=alpha", nil)
	rr := httptest.NewRecorder()
	handler.Results(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}

	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)

	resultsField := val.FieldByName("Results")
	if resultsField.Len() != 1 {
		t.Fatalf("expected 1 filtered result, got %d", resultsField.Len())
	}
	if !resultsField.MapIndex(reflect.ValueOf("AlphaModel")).IsValid() {
		t.Fatalf("expected AlphaModel to remain after search filtering")
	}
}

func TestResultsHandler_NormalizesNilMismatchedAndInvalidScores(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
		},
		Results: map[string]middleware.Result{
			"NilScores":   {Scores: nil},
			"ShortScores": {Scores: []int{200}}, // invalid + length mismatch
		},
	}
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

	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)

	resultsField := val.FieldByName("Results")
	nilScores := resultsField.MapIndex(reflect.ValueOf("NilScores"))
	if !nilScores.IsValid() {
		t.Fatalf("expected NilScores in template results")
	}
	shortScores := resultsField.MapIndex(reflect.ValueOf("ShortScores"))
	if !shortScores.IsValid() {
		t.Fatalf("expected ShortScores in template results")
	}

	nilScoresSlice := nilScores.FieldByName("Scores")
	if nilScoresSlice.Len() != 2 {
		t.Fatalf("expected NilScores Scores length 2, got %d", nilScoresSlice.Len())
	}

	shortScoresSlice := shortScores.FieldByName("Scores")
	if shortScoresSlice.Len() != 2 {
		t.Fatalf("expected ShortScores Scores length 2, got %d", shortScoresSlice.Len())
	}
	if shortScoresSlice.Index(0).Int() != 0 {
		t.Fatalf("expected invalid score to be normalized to 0, got %d", shortScoresSlice.Index(0).Int())
	}
}

func TestResultsHandler_GET_RenderError(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()
	ResultsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestResultsHandler_ZeroPrompts_DoesNotReturnNaN(t *testing.T) {
	restoreDir := changeToProjectRootResults(t)
	defer restoreDir()

	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add results but NO prompts - this causes NaN in pass percentage calculation
	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"TestModel": {Scores: []int{}}, // Empty scores because no prompts
	})

	req := httptest.NewRequest("GET", "/results", nil)
	rr := httptest.NewRecorder()

	ResultsHandler(rr, req)

	// When NaN occurs, the response body is incomplete/empty
	body := rr.Body.String()

	// Debug: print the body to see what we got
	t.Logf("Response body length: %d", len(body))
	t.Logf("Response body preview: %s", body[:min(500, len(body))])

	// The template will fail to render properly when NaN is in PassPercentages
	// Check that the response is actually complete and valid HTML
	if !strings.Contains(body, "</html>") {
		t.Error("response should contain complete HTML, but template rendering likely failed due to NaN")
	}
}

func TestResultsHandler_ProfileGroups_ColumnIndices_MultipleProfiles(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Math Prompt 1", Profile: "Math"},
			{Text: "Math Prompt 2", Profile: "Math"},
			{Text: "Science Prompt 1", Profile: "Science"},
			{Text: "Science Prompt 2", Profile: "Science"},
			{Text: "Writing Prompt 1", Profile: "Writing"},
		},
		Profiles: []middleware.Profile{
			{Name: "Math", Description: "Math problems"},
			{Name: "Science", Description: "Science questions"},
			{Name: "Writing", Description: "Writing tasks"},
		},
		Results: map[string]middleware.Result{
			"Model1": {Scores: []int{100, 100, 100, 100, 100}},
		},
	}
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

	data := renderer.RenderCalls[0].Data
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Struct {
		t.Fatalf("expected struct template data, got %T", data)
	}

	field := val.FieldByName("ProfileGroups")
	if !field.IsValid() {
		t.Fatalf("expected ProfileGroups field on template data")
	}

	// Verify we have 3 profile groups
	if field.Len() != 3 {
		t.Fatalf("expected 3 profile groups, got %d", field.Len())
	}

	// Check Math profile (first 2 prompts: indices 0-1)
	mathGroup := getProfileGroupByName(field, "Math")
	if mathGroup == nil {
		t.Fatal("Math profile group not found")
	}
	if mathGroup.StartCol != 0 {
		t.Errorf("expected Math StartCol=0, got %d", mathGroup.StartCol)
	}
	if mathGroup.EndCol != 1 {
		t.Errorf("expected Math EndCol=1, got %d", mathGroup.EndCol)
	}

	// Check Science profile (next 2 prompts: indices 2-3)
	scienceGroup := getProfileGroupByName(field, "Science")
	if scienceGroup == nil {
		t.Fatal("Science profile group not found")
	}
	if scienceGroup.StartCol != 2 {
		t.Errorf("expected Science StartCol=2, got %d", scienceGroup.StartCol)
	}
	if scienceGroup.EndCol != 3 {
		t.Errorf("expected Science EndCol=3, got %d", scienceGroup.EndCol)
	}

	// Check Writing profile (last prompt: index 4)
	writingGroup := getProfileGroupByName(field, "Writing")
	if writingGroup == nil {
		t.Fatal("Writing profile group not found")
	}
	if writingGroup.StartCol != 4 {
		t.Errorf("expected Writing StartCol=4, got %d", writingGroup.StartCol)
	}
	if writingGroup.EndCol != 4 {
		t.Errorf("expected Writing EndCol=4, got %d", writingGroup.EndCol)
	}
}

// Helper function to get profile group by name
func getProfileGroupByName(field reflect.Value, name string) *middleware.ProfileGroup {
	for i := 0; i < field.Len(); i++ {
		pg, ok := field.Index(i).Interface().(*middleware.ProfileGroup)
		if !ok {
			continue
		}
		if pg.Name == name {
			return pg
		}
	}
	return nil
}
