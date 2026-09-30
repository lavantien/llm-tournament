package middleware

import (
	"reflect"
	"testing"
)

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
		want  []int
	}{
		{
			name:  "empty map",
			input: map[string]Result{},
			want:  nil,
		},
		{
			name: "valid scores",
			input: map[string]Result{
				"Model": {Scores: []int{80, 60, 40}},
			},
			want: []int{80, 60, 40},
		},
		{
			name: "nil scores converted to empty",
			input: map[string]Result{
				"Model": {Scores: nil},
			},
			want: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MigrateResults(tt.input)
			if got == nil {
				t.Fatal("expected non-nil result")
			}
			if !reflect.DeepEqual(got["Model"].Scores, tt.want) {
				t.Errorf("migrated scores = %#v, want %#v", got["Model"].Scores, tt.want)
			}
		})
	}
}

func TestMigrateResults_ScoreClamping(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Out-of-range scores are zeroed, not capped at the boundary.
	high := MigrateResults(map[string]Result{
		"Model": {Scores: []int{150, 200}},
	})
	if got := high["Model"].Scores; len(got) != 2 || got[0] != 0 || got[1] != 0 {
		t.Errorf("expected [150 200] to migrate to [0 0], got %v", got)
	}

	negative := MigrateResults(map[string]Result{
		"Model": {Scores: []int{-10, 60}},
	})
	if got := negative["Model"].Scores; len(got) != 2 || got[0] != 0 || got[1] != 60 {
		t.Errorf("expected [-10 60] to migrate to [0 60], got %v", got)
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

func TestMigrateResults_NilScores(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Add prompts so we know the expected length
	_ = WritePrompts([]Prompt{{Text: "Prompt 1"}, {Text: "Prompt 2"}})

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
	_ = WritePrompts([]Prompt{{Text: "P1"}, {Text: "P2"}, {Text: "P3"}})

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
