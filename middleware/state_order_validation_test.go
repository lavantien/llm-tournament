package middleware

import (
	"testing"
)

// displayOrdersByPromptText reads the raw display_order column per prompt
// text for the default suite, so tests can assert persisted ordering state
// directly instead of relying on ReadPromptSuite tie-breaking.
func displayOrdersByPromptText(t *testing.T) map[string]int {
	t.Helper()

	rows, err := db.Query("SELECT text, display_order FROM prompts WHERE suite_id = (SELECT id FROM suites WHERE name = 'default')")
	if err != nil {
		t.Fatalf("failed to query display orders: %v", err)
	}
	defer func() { _ = rows.Close() }()

	orders := make(map[string]int)
	for rows.Next() {
		var text string
		var order int
		if err := rows.Scan(&text, &order); err != nil {
			t.Fatalf("failed to scan display order: %v", err)
		}
		orders[text] = order
	}
	return orders
}

func writeThreeOrderedPrompts(t *testing.T) {
	t.Helper()

	if err := WritePromptSuite("default", []Prompt{
		{Text: "P1"},
		{Text: "P2"},
		{Text: "P3"},
	}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
}

func assertDisplayOrdersUnchanged(t *testing.T, orders map[string]int) {
	t.Helper()

	expected := map[string]int{"P1": 0, "P2": 1, "P3": 2}
	for text, want := range expected {
		if got := orders[text]; got != want {
			t.Errorf("prompt %q display_order = %d, want unchanged %d", text, got, want)
		}
	}
}

func TestUpdatePromptsOrder_RejectsDuplicateIndex(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	writeThreeOrderedPrompts(t)

	// Index 0 appears twice, so two prompts would end up sharing a
	// display_order and one position would be overwritten.
	UpdatePromptsOrder([]int{0, 1, 0})

	assertDisplayOrdersUnchanged(t, displayOrdersByPromptText(t))
}

func TestUpdatePromptsOrder_RejectsOmittedIndex(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	writeThreeOrderedPrompts(t)

	// Index 1 is never referenced; index 2 is applied twice.
	UpdatePromptsOrder([]int{2, 0, 2})

	assertDisplayOrdersUnchanged(t, displayOrdersByPromptText(t))
}

func TestUpdatePromptsOrder_ValidPermutationStillApplies(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	writeThreeOrderedPrompts(t)

	UpdatePromptsOrder([]int{2, 0, 1})

	orders := displayOrdersByPromptText(t)
	expected := map[string]int{"P3": 0, "P1": 1, "P2": 2}
	for text, want := range expected {
		if got := orders[text]; got != want {
			t.Errorf("prompt %q display_order = %d, want %d", text, got, want)
		}
	}

	prompts := ReadPrompts()
	if len(prompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(prompts))
	}
	expectedSequence := []string{"P3", "P1", "P2"}
	for i, want := range expectedSequence {
		if prompts[i].Text != want {
			t.Errorf("prompts[%d] = %q, want %q", i, prompts[i].Text, want)
		}
	}
}
