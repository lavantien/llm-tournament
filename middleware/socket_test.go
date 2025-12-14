package middleware

import (
	"math"
	"testing"
)

func TestCalculatePassPercentages(t *testing.T) {
	tests := []struct {
		name        string
		results     map[string]Result
		promptCount int
		want        map[string]float64
	}{
		{
			name:        "empty results",
			results:     map[string]Result{},
			promptCount: 0,
			want:        map[string]float64{},
		},
		{
			name: "all zeros",
			results: map[string]Result{
				"Model A": {Scores: []int{0, 0, 0}},
			},
			promptCount: 3,
			want: map[string]float64{
				"Model A": 0,
			},
		},
		{
			name: "all 100s",
			results: map[string]Result{
				"Model A": {Scores: []int{100, 100, 100}},
			},
			promptCount: 3,
			want: map[string]float64{
				"Model A": 100,
			},
		},
		{
			name: "mixed scores",
			results: map[string]Result{
				"Model A": {Scores: []int{100, 50, 0}},
			},
			promptCount: 3,
			want: map[string]float64{
				"Model A": 50, // 150/300 * 100 = 50
			},
		},
		{
			name: "multiple models",
			results: map[string]Result{
				"Model A": {Scores: []int{100, 100}},
				"Model B": {Scores: []int{50, 50}},
			},
			promptCount: 2,
			want: map[string]float64{
				"Model A": 100,
				"Model B": 50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculatePassPercentages(tt.results, tt.promptCount)

			if len(got) != len(tt.want) {
				t.Errorf("expected %d results, got %d", len(tt.want), len(got))
			}

			for model, expected := range tt.want {
				if got[model] != expected {
					t.Errorf("model %q: expected %.2f, got %.2f", model, expected, got[model])
				}
			}
		})
	}
}

func TestCalculatePassPercentages_ZeroPrompts(t *testing.T) {
	results := map[string]Result{
		"Model A": {Scores: []int{}},
	}

	got := calculatePassPercentages(results, 0)

	// With zero prompts, percentage is NaN due to division by zero
	if !math.IsNaN(got["Model A"]) {
		t.Errorf("expected NaN for zero prompts, got %.2f", got["Model A"])
	}
}

func TestPromptsToStringArray(t *testing.T) {
	tests := []struct {
		name    string
		prompts []Prompt
		want    []string
	}{
		{
			name:    "empty prompts",
			prompts: []Prompt{},
			want:    []string{},
		},
		{
			name: "single prompt",
			prompts: []Prompt{
				{Text: "Prompt 1"},
			},
			want: []string{"Prompt 1"},
		},
		{
			name: "multiple prompts",
			prompts: []Prompt{
				{Text: "Prompt 1"},
				{Text: "Prompt 2"},
				{Text: "Prompt 3"},
			},
			want: []string{"Prompt 1", "Prompt 2", "Prompt 3"},
		},
		{
			name: "prompts with solutions and profiles",
			prompts: []Prompt{
				{Text: "Q1", Solution: "A1", Profile: "P1"},
				{Text: "Q2", Solution: "A2", Profile: "P2"},
			},
			want: []string{"Q1", "Q2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := promptsToStringArray(tt.prompts)

			if len(got) != len(tt.want) {
				t.Errorf("expected %d strings, got %d", len(tt.want), len(got))
			}

			for i, expected := range tt.want {
				if got[i] != expected {
					t.Errorf("index %d: expected %q, got %q", i, expected, got[i])
				}
			}
		})
	}
}

func TestProfileGroup_Struct(t *testing.T) {
	pg := ProfileGroup{
		ID:       "1",
		Name:     "Test Profile",
		StartCol: 0,
		EndCol:   5,
		Color:    "hsl(137, 70%, 50%)",
	}

	if pg.ID != "1" {
		t.Errorf("expected ID '1', got %q", pg.ID)
	}
	if pg.Name != "Test Profile" {
		t.Errorf("expected Name 'Test Profile', got %q", pg.Name)
	}
	if pg.StartCol != 0 {
		t.Errorf("expected StartCol 0, got %d", pg.StartCol)
	}
	if pg.EndCol != 5 {
		t.Errorf("expected EndCol 5, got %d", pg.EndCol)
	}
	if pg.Color != "hsl(137, 70%, 50%)" {
		t.Errorf("expected Color 'hsl(137, 70%%, 50%%)', got %q", pg.Color)
	}
}
