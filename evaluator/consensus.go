package evaluator

import "math"

// CalculateConsensusScore calculates weighted average score by confidence
func CalculateConsensusScore(results []JudgeResult) int {
	if len(results) == 0 {
		return 0
	}

	// Filter out invalid results
	validResults := make([]JudgeResult, 0)
	for _, r := range results {
		if r.Confidence > 0 && r.Score >= 0 && r.Score <= 100 {
			validResults = append(validResults, r)
		}
	}

	if len(validResults) == 0 {
		return 0
	}

	// Calculate weighted average
	// Note: totalWeight is always > 0 since validResults only contains items with Confidence > 0
	weightedSum := 0.0
	totalWeight := 0.0

	for _, result := range validResults {
		weightedSum += float64(result.Score) * result.Confidence
		totalWeight += result.Confidence
	}

	return int(math.Round(weightedSum / totalWeight))
}

// RoundToValidScore rounds score to nearest valid value [0, 20, 40, 60, 80, 100]
func RoundToValidScore(score int) int {
	validScores := []int{0, 20, 40, 60, 80, 100}

	// Find nearest valid score
	minDiff := 100
	nearest := 0

	for _, validScore := range validScores {
		diff := abs(score - validScore)
		if diff < minDiff {
			minDiff = diff
			nearest = validScore
		}
	}

	return nearest
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
