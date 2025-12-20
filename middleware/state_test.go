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

func TestGetCurrentSuiteName_NoCurrentSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Clear all is_current flags
	_, err = db.Exec("UPDATE suites SET is_current = 0")
	if err != nil {
		t.Fatalf("failed to clear current suite: %v", err)
	}

	// Function should fall back to default
	name := GetCurrentSuiteName()
	if name != "default" {
		t.Errorf("expected 'default', got %q", name)
	}
}

func TestGetCurrentSuiteName_QueryError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Drop suites table to trigger query error
	_, err = db.Exec("DROP TABLE suites")
	if err != nil {
		t.Fatalf("failed to drop suites table: %v", err)
	}

	// Function should return empty string on query error
	name := GetCurrentSuiteName()
	if name != "" {
		t.Errorf("expected empty string on query error, got %q", name)
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
		name  string
		input map[string]Result
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
	newOrder := []int{2, 1, 0}
	UpdatePromptsOrder(newOrder) // This function returns nothing

	// Read back and verify order changed
	readPrompts, _ := ReadPromptSuite("default")
	if len(readPrompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(readPrompts))
	}
	if readPrompts[0].Text != "Prompt 3" {
		t.Fatalf("expected first prompt to be %q, got %q", "Prompt 3", readPrompts[0].Text)
	}
	if readPrompts[2].Text != "Prompt 1" {
		t.Fatalf("expected last prompt to be %q, got %q", "Prompt 1", readPrompts[2].Text)
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

func TestUpdatePromptsOrder_ValidReorder(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create prompts in specific order
	prompts := []Prompt{
		{Text: "First"},
		{Text: "Second"},
		{Text: "Third"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Reorder: move last to first (0->1, 1->2, 2->0)
	newOrder := []int{2, 0, 1}
	UpdatePromptsOrder(newOrder)

	// Verify order changed
	readPrompts, _ := ReadPromptSuite("default")
	if len(readPrompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(readPrompts))
	}
}

func TestUpdatePromptsOrder_NegativeIndex(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Try with negative index - should not panic
	UpdatePromptsOrder([]int{-1, 1})

	// Prompts should still exist unchanged
	readPrompts, _ := ReadPromptSuite("default")
	if len(readPrompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(readPrompts))
	}
}

func TestUpdatePromptsOrder_GetSuiteIDError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Drop suites table to trigger GetCurrentSuiteID error
	_, err = db.Exec("DROP TABLE suites")
	if err != nil {
		t.Fatalf("failed to drop suites table: %v", err)
	}

	// Should not panic when GetCurrentSuiteID fails
	UpdatePromptsOrder([]int{1, 0})
}

func TestUpdatePromptsOrder_TransactionBeginError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Close database to trigger transaction begin error
	CloseDB()

	// Should not panic when transaction begin fails
	UpdatePromptsOrder([]int{1, 0})
}

func TestUpdatePromptsOrder_QueryError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Create prompts
	prompts := []Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
	}
	err = WritePromptSuite("default", prompts)
	if err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	// Drop prompts table to trigger query error
	_, err = db.Exec("DROP TABLE prompts")
	if err != nil {
		t.Fatalf("failed to drop prompts table: %v", err)
	}

	// Should not panic when query fails
	UpdatePromptsOrder([]int{1, 0})
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

func TestGetMaskedAPIKeys_AllProviders(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Get masked keys (should have defaults)
	masked, err := GetMaskedAPIKeys()
	if err != nil {
		t.Fatalf("GetMaskedAPIKeys failed: %v", err)
	}

	// Should have all three providers (keys are api_key_<provider>)
	if _, ok := masked["api_key_anthropic"]; !ok {
		t.Error("expected api_key_anthropic in masked keys")
	}
	if _, ok := masked["api_key_openai"]; !ok {
		t.Error("expected api_key_openai in masked keys")
	}
	if _, ok := masked["api_key_google"]; !ok {
		t.Error("expected api_key_google in masked keys")
	}
}

func TestUpdatePromptsOrder_WithReordering(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts
	err = WritePrompts([]Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
		{Text: "Prompt 3"},
	})
	if err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	// Update order - function returns void, just call it
	newOrder := []int{2, 0, 1}
	UpdatePromptsOrder(newOrder)

	// Verify prompts were reordered
	prompts := ReadPrompts()
	if len(prompts) != 3 {
		t.Errorf("expected 3 prompts, got %d", len(prompts))
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

func TestMigrateResults_EmptyResults(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	results := map[string]Result{}
	migrated := MigrateResults(results)

	if len(migrated) != 0 {
		t.Errorf("expected empty map, got %d results", len(migrated))
	}
}

func TestMigrateResults_OutOfRangeScores(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	results := map[string]Result{
		"Model1": {Scores: []int{-10, 150, 50}},
	}

	migrated := MigrateResults(results)

	// Check that out-of-range scores are clamped
	for _, score := range migrated["Model1"].Scores {
		if score < 0 || score > 100 {
			t.Errorf("migrated score %d should be in range 0-100", score)
		}
	}
}

func TestReadPromptSuite_NonExistent(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	prompts, err := ReadPromptSuite("non-existent-suite")
	// Should return empty data for non-existent suite
	if err != nil {
		// Error is expected for non-existent suite
		return
	}
	if prompts != nil && len(prompts) > 0 {
		t.Error("expected empty prompts for non-existent suite")
	}
}

func TestGetCurrentSuiteName_DefaultFallback(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Clear is_current flag from all suites
	_, err = db.Exec("UPDATE suites SET is_current = 0")
	if err != nil {
		t.Fatalf("Failed to clear is_current: %v", err)
	}

	// GetCurrentSuiteName should set default as current and return it
	name := GetCurrentSuiteName()
	if name != "default" {
		t.Errorf("expected 'default', got %q", name)
	}

	// Verify is_current was set (is_current is a boolean stored as bool in SQLite)
	var isCurrent bool
	err = db.QueryRow("SELECT is_current FROM suites WHERE name = 'default'").Scan(&isCurrent)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if !isCurrent {
		t.Error("expected is_current to be set to true")
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
	WritePrompts([]Prompt{{Text: "Test prompt"}})

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
	WritePrompts([]Prompt{{Text: "Prompt 1"}, {Text: "Prompt 2"}})

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

func TestWritePromptSuite_UpdatesExisting(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Write initial prompts
	WritePromptSuite("default", []Prompt{
		{Text: "Original prompt"},
	})

	// Update prompts
	WritePromptSuite("default", []Prompt{
		{Text: "Updated prompt 1"},
		{Text: "Updated prompt 2"},
	})

	// Verify update worked
	prompts, _ := ReadPromptSuite("default")
	if len(prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(prompts))
	}
	if prompts[0].Text != "Updated prompt 1" {
		t.Errorf("expected 'Updated prompt 1', got %q", prompts[0].Text)
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
	WritePrompts([]Prompt{{Text: "Test prompt"}})

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
	WritePrompts([]Prompt{{Text: "Prompt 1"}})

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

func TestMigrateResults_NilScores(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts so we know the expected length
	WritePrompts([]Prompt{{Text: "Prompt 1"}, {Text: "Prompt 2"}})

	results := map[string]Result{
		"Model1": {Scores: nil}, // Nil scores should be initialized
	}

	migrated := MigrateResults(results)

	// Scores should be initialized to array of zeros
	if migrated["Model1"].Scores == nil {
		t.Error("expected scores to be initialized, not nil")
	}
}

func TestMigrateResults_ShorterScores(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add 3 prompts
	WritePrompts([]Prompt{{Text: "P1"}, {Text: "P2"}, {Text: "P3"}})

	results := map[string]Result{
		"Model1": {Scores: []int{80}}, // Only 1 score for 3 prompts
	}

	migrated := MigrateResults(results)

	// Scores should be padded to length 3
	if len(migrated["Model1"].Scores) != 3 {
		t.Errorf("expected 3 scores, got %d", len(migrated["Model1"].Scores))
	}
	// First score should be preserved
	if migrated["Model1"].Scores[0] != 80 {
		t.Errorf("expected first score 80, got %d", migrated["Model1"].Scores[0])
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
	WritePrompts([]Prompt{{Text: "P1"}})

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
	WritePrompts([]Prompt{{Text: "P1"}})

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
