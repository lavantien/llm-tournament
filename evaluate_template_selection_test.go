package main

import (
	"os"
	"strings"
	"testing"
)

func TestEvaluateTemplate_UsesDataScoreForInitialSelection(t *testing.T) {
	t.Helper()

	b, err := os.ReadFile("templates/evaluate.html")
	if err != nil {
		t.Fatalf("read templates/evaluate.html: %v", err)
	}
	s := string(b)

	if strings.Contains(s, ".onclick.toString()") {
		t.Fatalf("templates/evaluate.html should not rely on onclick.toString() for selection")
	}
	if !strings.Contains(s, `.score-button[data-score="${currentScore}"]`) {
		t.Fatalf("templates/evaluate.html should select initial score via data-score selector")
	}
}
