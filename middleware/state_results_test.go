package middleware

import (
	"testing"
)

func TestReadProfileSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Initially should return empty slice
	profiles, err := ReadProfileSuite("default")
	if err != nil {
		t.Fatalf("ReadProfileSuite failed: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestWriteProfileSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	profiles := []Profile{
		{Name: "Profile A", Description: "Description A"},
		{Name: "Profile B", Description: "Description B"},
	}

	err = WriteProfileSuite("default", profiles)
	if err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}

	// Read back and verify
	readProfiles, err := ReadProfileSuite("default")
	if err != nil {
		t.Fatalf("ReadProfileSuite failed: %v", err)
	}
	if len(readProfiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(readProfiles))
	}
	if readProfiles[0].Name != "Profile A" {
		t.Errorf("expected 'Profile A', got %q", readProfiles[0].Name)
	}
}

func TestWriteProfileSuite_Overwrite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Write initial profiles
	err = WriteProfileSuite("default", []Profile{
		{Name: "Old Profile"},
	})
	if err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}

	// Overwrite with new profiles
	err = WriteProfileSuite("default", []Profile{
		{Name: "New Profile 1"},
		{Name: "New Profile 2"},
	})
	if err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}

	// Verify old profiles are gone
	profiles, _ := ReadProfileSuite("default")
	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].Name != "New Profile 1" {
		t.Errorf("expected 'New Profile 1', got %q", profiles[0].Name)
	}
}

func TestReadResults_Empty(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	results := ReadResults()
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestWriteResults(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// First write prompts so scores have targets
	prompts := []Prompt{
		{Text: "Prompt 1", Solution: "Solution 1"},
		{Text: "Prompt 2", Solution: "Solution 2"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Write results
	results := map[string]Result{
		"Model A": {Scores: []int{80, 60}},
		"Model B": {Scores: []int{100, 40}},
	}

	err = WriteResults("default", results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Read back and verify
	readResults := ReadResults()
	if len(readResults) != 2 {
		t.Errorf("expected 2 models, got %d", len(readResults))
	}
	if readResults["Model A"].Scores[0] != 80 {
		t.Errorf("expected Model A score 80, got %d", readResults["Model A"].Scores[0])
	}
}

func TestReadProfiles(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Initially should return empty
	profiles := ReadProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}

	// Write some profiles
	err = WriteProfileSuite("default", []Profile{
		{Name: "Profile 1", Description: "Desc 1"},
		{Name: "Profile 2", Description: "Desc 2"},
	})
	if err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}

	// Read back
	profiles = ReadProfiles()
	if len(profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestWriteProfiles(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	profiles := []Profile{
		{Name: "Test Profile", Description: "Test Description"},
	}

	err = WriteProfiles(profiles)
	if err != nil {
		t.Fatalf("WriteProfiles failed: %v", err)
	}

	// Read back
	readProfiles := ReadProfiles()
	if len(readProfiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(readProfiles))
	}
}

func TestWriteResults_OverwriteExisting(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// First write prompts
	prompts := []Prompt{
		{Text: "Prompt 1", Solution: "Solution 1"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Write initial results
	results := map[string]Result{
		"Model A": {Scores: []int{60}},
	}
	err = WriteResults("default", results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Overwrite with new results
	newResults := map[string]Result{
		"Model A": {Scores: []int{80}},
		"Model B": {Scores: []int{100}},
	}
	err = WriteResults("default", newResults)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Verify new results
	readResults := ReadResults()
	if len(readResults) != 2 {
		t.Errorf("expected 2 models, got %d", len(readResults))
	}
	if readResults["Model A"].Scores[0] != 80 {
		t.Errorf("expected Model A score 80, got %d", readResults["Model A"].Scores[0])
	}
}

func TestWriteResults_NewModel(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts first
	prompts := []Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Write results for a new model
	results := map[string]Result{
		"NewModel": {Scores: []int{80, 100}},
	}
	err = WriteResults("default", results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Read back
	readResults := ReadResults()
	if len(readResults) != 1 {
		t.Errorf("expected 1 model, got %d", len(readResults))
	}
	if readResults["NewModel"].Scores[0] != 80 {
		t.Errorf("expected score 80, got %d", readResults["NewModel"].Scores[0])
	}
}

func TestWriteResults_UpdateExistingModel(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Write initial results
	results := map[string]Result{
		"ExistingModel": {Scores: []int{40}},
	}
	err = WriteResults("default", results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Update results
	results["ExistingModel"] = Result{Scores: []int{80}}
	err = WriteResults("default", results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Read back
	readResults := ReadResults()
	if readResults["ExistingModel"].Scores[0] != 80 {
		t.Errorf("expected updated score 80, got %d", readResults["ExistingModel"].Scores[0])
	}
}

func TestReadResults_ScoreMismatch(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Add model with only 1 score (less than prompts)
	results := map[string]Result{
		"TestModel": {Scores: []int{60}},
	}
	err = WriteResults("default", results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Read back - should pad with zeros
	readResults := ReadResults()
	if len(readResults["TestModel"].Scores) != 2 {
		t.Errorf("expected 2 scores (padded), got %d", len(readResults["TestModel"].Scores))
	}
}

func TestWriteResults_MultipleModels(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts
	err = WritePrompts([]Prompt{{Text: "Prompt 1"}, {Text: "Prompt 2"}})
	if err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	// Write results for multiple models
	results := map[string]Result{
		"Model A": {Scores: []int{80, 60}},
		"Model B": {Scores: []int{100, 40}},
		"Model C": {Scores: []int{60, 80}},
	}

	err = WriteResults("default", results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Read back
	readResults := ReadResults()
	if len(readResults) != 3 {
		t.Errorf("expected 3 models, got %d", len(readResults))
	}
}

func TestWriteResults_UpdateExisting(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	err = WritePrompts([]Prompt{{Text: "Prompt 1"}})
	if err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	// Write initial results
	err = WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{50}},
	})
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Update results
	err = WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{80}},
	})
	if err != nil {
		t.Fatalf("WriteResults update failed: %v", err)
	}

	// Verify update
	results := ReadResults()
	if results["Model1"].Scores[0] != 80 {
		t.Errorf("expected score 80, got %d", results["Model1"].Scores[0])
	}
}

func TestWriteProfileSuite_Success(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	profiles := []Profile{
		{Name: "Profile 1", Description: "Desc 1"},
		{Name: "Profile 2", Description: "Desc 2"},
	}

	err = WriteProfileSuite("test-suite", profiles)
	if err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}

	// Read back
	readProfiles, err := ReadProfileSuite("test-suite")
	if err != nil {
		t.Fatalf("ReadProfileSuite failed: %v", err)
	}
	if len(readProfiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(readProfiles))
	}
}

func TestReadResults_WithMissingScores(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add 3 prompts
	err = WritePrompts([]Prompt{{Text: "P1"}, {Text: "P2"}, {Text: "P3"}})
	if err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	// Write results with only 1 score
	err = WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{50}},
	})
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Read back should pad with zeros
	results := ReadResults()
	if len(results["Model1"].Scores) != 3 {
		t.Errorf("expected 3 scores (padded), got %d", len(results["Model1"].Scores))
	}
}

func TestReadResults_EmptySuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Reading results from fresh suite should return empty map
	results := ReadResults()
	if results == nil {
		t.Error("expected non-nil results map")
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d entries", len(results))
	}
}

func TestReadResults_WithPromptsNoScores(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts
	_ = WritePrompts([]Prompt{{Text: "Test prompt"}})

	// Results should still be empty (no models/scores)
	results := ReadResults()
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d entries", len(results))
	}
}

func TestWriteResults_WithMultipleModels(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts first
	_ = WritePrompts([]Prompt{{Text: "Prompt 1"}, {Text: "Prompt 2"}})

	// Write results for multiple models
	results := map[string]Result{
		"Model1": {Scores: []int{80, 60}},
		"Model2": {Scores: []int{100, 40}},
		"Model3": {Scores: []int{20, 0}},
	}
	err = WriteResults("default", results)
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Verify all models have results
	readResults := ReadResults()
	if len(readResults) != 3 {
		t.Errorf("expected 3 models, got %d", len(readResults))
	}
}

func TestReadResults_AfterWriteResults(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts
	_ = WritePrompts([]Prompt{{Text: "Test prompt"}})

	// Write results
	err = WriteResults("default", map[string]Result{
		"TestModel": {Scores: []int{80}},
	})
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Read results
	results := ReadResults()
	if len(results) != 1 {
		t.Errorf("expected 1 model, got %d", len(results))
	}
	if _, exists := results["TestModel"]; !exists {
		t.Error("expected TestModel to exist")
	}
}

func TestWriteResults_DeleteModel(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts
	_ = WritePrompts([]Prompt{{Text: "Prompt 1"}})

	// Write initial results with two models
	err = WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{80}},
		"Model2": {Scores: []int{60}},
	})
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Verify both models exist
	results := ReadResults()
	if len(results) != 2 {
		t.Errorf("expected 2 models initially, got %d", len(results))
	}

	// Write results with only one model (should delete Model2)
	err = WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{100}},
	})
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Verify Model2 was deleted
	results = ReadResults()
	if len(results) != 1 {
		t.Errorf("expected 1 model after deletion, got %d", len(results))
	}
	if _, exists := results["Model2"]; exists {
		t.Error("Model2 should have been deleted")
	}
}

func TestWriteResults_EmptyScores(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts
	_ = WritePrompts([]Prompt{{Text: "P1"}})

	// Write results with empty scores
	err = WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{}},
	})
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Should succeed without panic
	results := ReadResults()
	if _, exists := results["Model1"]; !exists {
		t.Error("Model1 should exist even with empty scores")
	}
}

func TestWriteResults_CreateNewModel(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts
	_ = WritePrompts([]Prompt{{Text: "P1"}})

	// Write results with a new model (should create it)
	err = WriteResults("default", map[string]Result{
		"NewModel": {Scores: []int{75}},
	})
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Verify model was created
	results := ReadResults()
	if _, exists := results["NewModel"]; !exists {
		t.Error("NewModel should have been created")
	}
	if results["NewModel"].Scores[0] != 75 {
		t.Errorf("expected score 75, got %d", results["NewModel"].Scores[0])
	}
}
