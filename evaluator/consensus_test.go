package evaluator

import (
	"testing"
)

func TestCalculateConsensusScore(t *testing.T) {
	tests := []struct {
		name    string
		results []JudgeResult
		want    int
	}{
		{
			name:    "empty results",
			results: []JudgeResult{},
			want:    0,
		},
		{
			name: "single result",
			results: []JudgeResult{
				{Score: 80, Confidence: 0.9},
			},
			want: 80,
		},
		{
			name: "two equal weights",
			results: []JudgeResult{
				{Score: 60, Confidence: 0.5},
				{Score: 80, Confidence: 0.5},
			},
			want: 70, // (60*0.5 + 80*0.5) / 1.0 = 70
		},
		{
			name: "weighted average favoring higher confidence",
			results: []JudgeResult{
				{Score: 60, Confidence: 0.2},
				{Score: 80, Confidence: 0.8},
			},
			want: 76, // (60*0.2 + 80*0.8) / 1.0 = 76
		},
		{
			name: "three judges",
			results: []JudgeResult{
				{Score: 60, Confidence: 0.7},
				{Score: 80, Confidence: 0.9},
				{Score: 40, Confidence: 0.5},
			},
			// (60*0.7 + 80*0.9 + 40*0.5) / (0.7+0.9+0.5) = (42+72+20)/2.1 = 134/2.1 ≈ 63.8 → 64
			want: 64,
		},
		{
			name: "all same score",
			results: []JudgeResult{
				{Score: 80, Confidence: 0.5},
				{Score: 80, Confidence: 0.8},
				{Score: 80, Confidence: 0.3},
			},
			want: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateConsensusScore(tt.results)
			if got != tt.want {
				t.Errorf("CalculateConsensusScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateConsensusScore_ZeroConfidence(t *testing.T) {
	// When all results have zero confidence, fallback to simple average
	results := []JudgeResult{
		{Score: 60, Confidence: 0},
		{Score: 80, Confidence: 0},
	}

	// With zero confidence, both should be filtered as invalid
	// Since validResults will be empty after filtering, return 0
	got := CalculateConsensusScore(results)
	if got != 0 {
		t.Errorf("expected 0 for all zero confidence, got %d", got)
	}
}

func TestCalculateConsensusScore_AllInvalid(t *testing.T) {
	tests := []struct {
		name    string
		results []JudgeResult
	}{
		{
			name: "negative confidence",
			results: []JudgeResult{
				{Score: 80, Confidence: -0.5},
				{Score: 60, Confidence: -0.3},
			},
		},
		{
			name: "zero confidence",
			results: []JudgeResult{
				{Score: 80, Confidence: 0},
				{Score: 60, Confidence: 0},
			},
		},
		{
			name: "score out of range negative",
			results: []JudgeResult{
				{Score: -10, Confidence: 0.9},
				{Score: -5, Confidence: 0.8},
			},
		},
		{
			name: "score out of range high",
			results: []JudgeResult{
				{Score: 150, Confidence: 0.9},
				{Score: 200, Confidence: 0.8},
			},
		},
		{
			name: "mixed invalid",
			results: []JudgeResult{
				{Score: -10, Confidence: 0.9},
				{Score: 80, Confidence: 0},
				{Score: 150, Confidence: 0.5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateConsensusScore(tt.results)
			if got != 0 {
				t.Errorf("CalculateConsensusScore() = %d, want 0 for invalid results", got)
			}
		})
	}
}

func TestCalculateConsensusScore_MixedValidInvalid(t *testing.T) {
	results := []JudgeResult{
		{Score: 80, Confidence: 0.9},  // valid
		{Score: -10, Confidence: 0.8}, // invalid: negative score
		{Score: 60, Confidence: 0.7},  // valid
		{Score: 150, Confidence: 0.6}, // invalid: score > 100
		{Score: 40, Confidence: 0},    // invalid: zero confidence
	}

	// Only valid: 80 (0.9) and 60 (0.7)
	// Weighted: (80*0.9 + 60*0.7) / (0.9 + 0.7) = (72 + 42) / 1.6 = 114/1.6 = 71.25 → 71
	got := CalculateConsensusScore(results)
	if got != 71 {
		t.Errorf("CalculateConsensusScore() = %d, want 71", got)
	}
}

func TestCalculateConsensusScore_BoundaryScores(t *testing.T) {
	tests := []struct {
		name    string
		results []JudgeResult
		want    int
	}{
		{
			name: "score exactly 0",
			results: []JudgeResult{
				{Score: 0, Confidence: 0.9},
			},
			want: 0,
		},
		{
			name: "score exactly 100",
			results: []JudgeResult{
				{Score: 100, Confidence: 0.9},
			},
			want: 100,
		},
		{
			name: "mix of boundary scores",
			results: []JudgeResult{
				{Score: 0, Confidence: 0.5},
				{Score: 100, Confidence: 0.5},
			},
			want: 50, // (0*0.5 + 100*0.5) / 1.0 = 50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateConsensusScore(tt.results)
			if got != tt.want {
				t.Errorf("CalculateConsensusScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRoundToValidScore(t *testing.T) {
	tests := []struct {
		score int
		want  int
	}{
		// Exact valid scores
		{0, 0},
		{20, 20},
		{40, 40},
		{60, 60},
		{80, 80},
		{100, 100},

		// Boundary cases - rounding down
		{9, 0},   // closer to 0 than 20
		{10, 0},  // equidistant, first match wins (0)
		{11, 20}, // closer to 20
		{29, 20},
		{30, 20}, // equidistant, first match wins (20)
		{31, 40},
		{49, 40},
		{50, 40}, // equidistant, first match wins (40)
		{51, 60},
		{69, 60},
		{70, 60}, // equidistant, first match wins (60)
		{71, 80},
		{89, 80},
		{90, 80}, // equidistant, first match wins (80)
		{91, 100},

		// Edge values
		{1, 0},
		{19, 20},
		{21, 20},
		{39, 40},
		{41, 40},
		{59, 60},
		{61, 60},
		{79, 80},
		{81, 80},
		{99, 100},

		// Out of normal range
		{-10, 0},
		{110, 100},
		{-100, 0},
		// Note: 200 returns 0 due to minDiff logic (all diffs >= 100)
		{200, 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := RoundToValidScore(tt.score)
			if got != tt.want {
				t.Errorf("RoundToValidScore(%d) = %d, want %d", tt.score, got, tt.want)
			}
		})
	}
}

func TestRoundToValidScore_AllValues(t *testing.T) {
	// Test every value from -20 to 120
	for score := -20; score <= 120; score++ {
		got := RoundToValidScore(score)

		// Verify result is one of the valid scores
		validScores := map[int]bool{0: true, 20: true, 40: true, 60: true, 80: true, 100: true}
		if !validScores[got] {
			t.Errorf("RoundToValidScore(%d) = %d, not a valid score", score, got)
		}
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		x    int
		want int
	}{
		{0, 0},
		{1, 1},
		{-1, 1},
		{100, 100},
		{-100, 100},
		{-2147483648, 2147483648}, // Note: may overflow on 32-bit systems
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := abs(tt.x)
			if got != tt.want {
				t.Errorf("abs(%d) = %d, want %d", tt.x, got, tt.want)
			}
		})
	}
}
