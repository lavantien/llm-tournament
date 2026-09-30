package handlers

import (
	"encoding/json"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func postPromptsOrder(t *testing.T, order []int) *httptest.ResponseRecorder {
	t.Helper()
	orderJSON, _ := json.Marshal(order)
	form := url.Values{}
	form.Add("order", string(orderJSON))
	req := httptest.NewRequest("POST", "/update_prompts_order", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	UpdatePromptsOrderHandler(rr, req)
	return rr
}

func writeABCOrderPrompts(t *testing.T) {
	t.Helper()
	suiteName := middleware.GetCurrentSuiteName()
	if err := middleware.WritePromptSuite(suiteName, []middleware.Prompt{
		{Text: "A"},
		{Text: "B"},
		{Text: "C"},
	}); err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}
}

func currentPromptTexts(t *testing.T) []string {
	t.Helper()
	prompts := middleware.ReadPrompts()
	texts := make([]string, 0, len(prompts))
	for _, p := range prompts {
		texts = append(texts, p.Text)
	}
	return texts
}

func TestUpdatePromptsOrderHandler_RejectsPartialOrder(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()
	writeABCOrderPrompts(t)

	// A filtered or searched prompt list can only post its rendered subset
	rr := postPromptsOrder(t, []int{0, 2})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for partial order, got %d", http.StatusBadRequest, rr.Code)
	}
	expected := []string{"A", "B", "C"}
	if got := currentPromptTexts(t); !reflect.DeepEqual(got, expected) {
		t.Errorf("expected order unchanged %v, got %v", expected, got)
	}
}

func TestUpdatePromptsOrderHandler_RejectsDuplicateIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()
	writeABCOrderPrompts(t)

	rr := postPromptsOrder(t, []int{0, 0, 1})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for duplicate index, got %d", http.StatusBadRequest, rr.Code)
	}
	expected := []string{"A", "B", "C"}
	if got := currentPromptTexts(t); !reflect.DeepEqual(got, expected) {
		t.Errorf("expected order unchanged %v, got %v", expected, got)
	}
}

func TestUpdatePromptsOrderHandler_AcceptsFullPermutation(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()
	writeABCOrderPrompts(t)

	rr := postPromptsOrder(t, []int{2, 0, 1})

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for a full permutation, got %d", http.StatusSeeOther, rr.Code)
	}
	expected := []string{"C", "A", "B"}
	if got := currentPromptTexts(t); !reflect.DeepEqual(got, expected) {
		t.Errorf("expected applied order %v, got %v", expected, got)
	}
}
