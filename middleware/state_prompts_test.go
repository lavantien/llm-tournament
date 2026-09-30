package middleware

import (
	"testing"
)

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

	// Verify the actual sequence, not just the count: order [2, 0, 1]
	// puts Third first, First second, and Second last.
	readPrompts, _ := ReadPromptSuite("default")
	if len(readPrompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(readPrompts))
	}
	expectedSequence := []string{"Third", "First", "Second"}
	for i, want := range expectedSequence {
		if readPrompts[i].Text != want {
			t.Errorf("readPrompts[%d] = %q, want %q", i, readPrompts[i].Text, want)
		}
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
	_ = CloseDB()

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

	// Verify the actual sequence, not just the count: order [2, 0, 1]
	// puts Prompt 3 first, Prompt 1 second, and Prompt 2 last.
	prompts := ReadPrompts()
	if len(prompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(prompts))
	}
	expectedSequence := []string{"Prompt 3", "Prompt 1", "Prompt 2"}
	for i, want := range expectedSequence {
		if prompts[i].Text != want {
			t.Errorf("prompts[%d] = %q, want %q", i, prompts[i].Text, want)
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

	// GetSuiteID auto-creates unknown suites, so reading a nonexistent
	// suite yields empty prompts without an error.
	prompts, err := ReadPromptSuite("non-existent-suite")
	if err != nil {
		t.Fatalf("expected no error for non-existent suite, got %v", err)
	}
	if len(prompts) != 0 {
		t.Errorf("expected 0 prompts for non-existent suite, got %d", len(prompts))
	}
	if !SuiteExists("non-existent-suite") {
		t.Error("expected the suite to be created by the lookup")
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
	_ = WritePromptSuite("default", []Prompt{
		{Text: "Original prompt"},
	})

	// Update prompts
	_ = WritePromptSuite("default", []Prompt{
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
