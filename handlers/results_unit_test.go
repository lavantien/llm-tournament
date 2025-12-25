package handlers

import (
	"math/rand"
	"testing"
)

func TestTierAlgorithm(t *testing.T) {
	numModels := 36
	numPrompts := 50
	numTiers := 12
	maxScore := numPrompts * 100
	rng := rand.New(rand.NewSource(42))

	tierCounts := make(map[int]int)
	scoreSums := make(map[int]int)

	for i := 0; i < numModels; i++ {
		tierIndex := (i * numTiers) / numModels
		if tierIndex >= numTiers {
			tierIndex = numTiers - 1
		}

		tierMinPercent := float64(tierIndex) / float64(numTiers)
		tierMaxPercent := float64(tierIndex+1) / float64(numTiers)
		tierMidPercent := (tierMinPercent + tierMaxPercent) / 2
		targetTotal := int(float64(maxScore) * tierMidPercent)

		tierCounts[tierIndex]++

		// Simulate score generation
		remaining := targetTotal
		totalScore := 0
		for j := 0; j < numPrompts; j++ {
			var score int
			if j == numPrompts-1 {
				score = clampToValidScore(remaining)
			} else {
				promptsRemaining := numPrompts - j - 1
				minScore := max(0, remaining-promptsRemaining*100)
				maxScoreVal := min(100, remaining)

				midScore := (minScore + maxScoreVal) / 2
				randomOffset := rng.Intn(21) - 10
				score = clampToValidScore(midScore + randomOffset)

				if score > maxScoreVal {
					score = clampToValidScore(maxScoreVal)
				}
				if score < minScore {
					score = clampToValidScore(minScore)
				}
			}

			totalScore += score
			remaining -= score
		}

		scoreSums[tierIndex] += totalScore

		t.Logf("Model %d: tier=%d, targetTotal=%d, actualTotal=%d, remaining=%d",
			i, tierIndex, targetTotal, totalScore, remaining)
	}

	t.Logf("Tier distribution: %v", tierCounts)
	t.Logf("Score sums by tier: %v", scoreSums)

	if len(tierCounts) < 10 {
		t.Errorf("Expected at least 10 tiers, got %d", len(tierCounts))
	}
}
