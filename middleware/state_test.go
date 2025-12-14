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

func TestReadPromptSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Initially should return empty slice
	prompts, err := ReadPromptSuite("default")
	if err != nil {
		t.Fatalf("ReadPromptSuite failed: %v", err)
	}
	if len(prompts) != 0 {
		t.Errorf("expected 0 prompts, got %d", len(prompts))
	}
}

func TestWritePromptSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	prompts := []Prompt{
		{Text: "Prompt 1", Solution: "Solution 1", Profile: ""},
		{Text: "Prompt 2", Solution: "Solution 2", Profile: ""},
	}

	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Read back and verify
	readPrompts, err := ReadPromptSuite("default")
	if err != nil {
		t.Fatalf("ReadPromptSuite failed: %v", err)
	}
	if len(readPrompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(readPrompts))
	}
	if readPrompts[0].Text != "Prompt 1" {
		t.Errorf("expected 'Prompt 1', got %q", readPrompts[0].Text)
	}
}

func TestWritePromptSuite_WithProfile(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// First create profiles
	err = WriteProfileSuite("default", []Profile{
		{Name: "Test Profile", Description: "Test"},
	})
	if err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}

	// Write prompts with profile reference
	prompts := []Prompt{
		{Text: "Prompt 1", Solution: "Solution 1", Profile: "Test Profile"},
	}

	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Read back and verify
	readPrompts, _ := ReadPromptSuite("default")
	if len(readPrompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(readPrompts))
	}
	if readPrompts[0].Profile != "Test Profile" {
		t.Errorf("expected profile 'Test Profile', got %q", readPrompts[0].Profile)
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

func TestGetCurrentSuiteName(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	name := GetCurrentSuiteName()
	if name != "default" {
		t.Errorf("expected 'default', got %q", name)
	}
}

func TestGetCurrentSuiteName_AfterSwitch(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Switch to a new suite
	err = SetCurrentSuite("test-suite")
	if err != nil {
		t.Fatalf("SetCurrentSuite failed: %v", err)
	}

	name := GetCurrentSuiteName()
	if name != "test-suite" {
		t.Errorf("expected 'test-suite', got %q", name)
	}
}

func TestSuiteExists(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if !SuiteExists("default") {
		t.Error("expected default suite to exist")
	}

	if SuiteExists("nonexistent") {
		t.Error("expected nonexistent suite to not exist")
	}
}

func TestListPromptSuites(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create additional suites
	GetSuiteID("suite-a")
	GetSuiteID("suite-b")

	suites, err := ListPromptSuites()
	if err != nil {
		t.Fatalf("ListPromptSuites failed: %v", err)
	}

	if len(suites) != 3 {
		t.Errorf("expected 3 suites, got %d", len(suites))
	}
}

func TestMigrateResults(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	tests := []struct {
		name   string
		input  map[string]Result
	}{
		{
			name:  "empty map",
			input: map[string]Result{},
		},
		{
			name: "valid scores",
			input: map[string]Result{
				"Model": {Scores: []int{80, 60, 40}},
			},
		},
		{
			name: "nil scores converted to empty",
			input: map[string]Result{
				"Model": {Scores: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MigrateResults(tt.input)
			if got == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestUpdatePromptsOrder(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create some prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
		{Text: "Prompt 3"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Reorder prompts (reverse order)
	newOrder := []int{3, 2, 1}
	UpdatePromptsOrder(newOrder) // This function returns nothing

	// Read back and verify order changed
	readPrompts, _ := ReadPromptSuite("default")
	if len(readPrompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(readPrompts))
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

func TestReadPrompts(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Initially should return empty
	prompts := ReadPrompts()
	if len(prompts) != 0 {
		t.Errorf("expected 0 prompts, got %d", len(prompts))
	}

	// Write some prompts
	err = WritePromptSuite("default", []Prompt{
		{Text: "Prompt 1", Solution: "Sol 1"},
		{Text: "Prompt 2", Solution: "Sol 2"},
	})
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Read back
	prompts = ReadPrompts()
	if len(prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(prompts))
	}
}

func TestWritePrompts(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	prompts := []Prompt{
		{Text: "Test Prompt", Solution: "Test Solution"},
	}

	err = WritePrompts(prompts)
	if err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	// Read back
	readPrompts := ReadPrompts()
	if len(readPrompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(readPrompts))
	}
}

func TestDeletePromptSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a new suite
	GetSuiteID("test-delete-suite")

	// Verify it exists
	if !SuiteExists("test-delete-suite") {
		t.Fatal("test-delete-suite should exist before deletion")
	}

	// Delete it
	err = DeletePromptSuite("test-delete-suite")
	if err != nil {
		t.Fatalf("DeletePromptSuite failed: %v", err)
	}

	// Verify it's gone
	if SuiteExists("test-delete-suite") {
		t.Error("test-delete-suite should not exist after deletion")
	}
}

func TestDeleteProfileSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a suite and add profiles
	GetSuiteID("profile-delete-suite")
	err = WriteProfileSuite("profile-delete-suite", []Profile{
		{Name: "Profile to Delete"},
	})
	if err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}

	// Delete it
	err = DeleteProfileSuite("profile-delete-suite")
	if err != nil {
		t.Fatalf("DeleteProfileSuite failed: %v", err)
	}
}

func TestListProfileSuites(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create additional suites
	GetSuiteID("profile-suite-a")
	GetSuiteID("profile-suite-b")

	suites, err := ListProfileSuites()
	if err != nil {
		t.Fatalf("ListProfileSuites failed: %v", err)
	}

	if len(suites) != 3 {
		t.Errorf("expected 3 suites, got %d", len(suites))
	}
}

func TestRenameSuiteFiles(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create a suite
	GetSuiteID("old-suite-name")
	if !SuiteExists("old-suite-name") {
		t.Fatal("old-suite-name should exist before rename")
	}

	// Rename it
	err = RenameSuiteFiles("old-suite-name", "new-suite-name")
	if err != nil {
		t.Fatalf("RenameSuiteFiles failed: %v", err)
	}

	// Verify old name is gone and new name exists
	if SuiteExists("old-suite-name") {
		t.Error("old-suite-name should not exist after rename")
	}
	if !SuiteExists("new-suite-name") {
		t.Error("new-suite-name should exist after rename")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{-1, 0, -1},
		{0, -1, -1},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.expected {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestWritePromptSuite_WithSolution(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	prompts := []Prompt{
		{Text: "Prompt with Solution", Solution: "This is the solution"},
		{Text: "Prompt without Solution", Solution: ""},
	}

	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Read back and verify
	readPrompts, _ := ReadPromptSuite("default")
	if len(readPrompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(readPrompts))
	}
	if readPrompts[0].Solution != "This is the solution" {
		t.Errorf("expected solution 'This is the solution', got %q", readPrompts[0].Solution)
	}
	if readPrompts[1].Solution != "" {
		t.Errorf("expected empty solution, got %q", readPrompts[1].Solution)
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

func TestMigrateResults_ScoreClamping(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Test that scores above 100 are clamped
	results := map[string]Result{
		"Model": {Scores: []int{150, 200}},
	}

	migrated := MigrateResults(results)
	if migrated["Model"].Scores[0] > 100 {
		t.Errorf("expected score <= 100, got %d", migrated["Model"].Scores[0])
	}
}

func TestUpdatePromptsOrder_InvalidOrder(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create some prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Try with invalid order (non-existent IDs)
	invalidOrder := []int{999, 998}
	UpdatePromptsOrder(invalidOrder) // Should not panic

	// Prompts should still exist
	readPrompts, _ := ReadPromptSuite("default")
	if len(readPrompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(readPrompts))
	}
}

func TestUpdatePromptsOrder_EmptyOrder(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create some prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Try with empty order
	UpdatePromptsOrder([]int{}) // Should not panic

	// Prompts should still exist
	readPrompts, _ := ReadPromptSuite("default")
	if len(readPrompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(readPrompts))
	}
}
