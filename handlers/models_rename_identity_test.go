package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"llm-tournament/middleware"

	_ "github.com/mattn/go-sqlite3"
)

// WriteResults reconciles models by name, so a rename that deletes the old
// model row and inserts a new one cascades away its model_responses. The
// rename must happen in place for saved responses to survive it.
func TestEditModelHandler_POST_RenamePreservesModelResponses(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	if err := middleware.WritePrompts([]middleware.Prompt{{Text: "p1"}}); err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}
	if err := middleware.WriteResults("default", map[string]middleware.Result{
		"OldModel": {Scores: []int{42}},
	}); err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	db := middleware.GetDB()
	var modelID, promptID int
	if err := db.QueryRow("SELECT id FROM models WHERE name = 'OldModel'").Scan(&modelID); err != nil {
		t.Fatalf("failed to read model id: %v", err)
	}
	if err := db.QueryRow("SELECT id FROM prompts WHERE text = 'p1'").Scan(&promptID); err != nil {
		t.Fatalf("failed to read prompt id: %v", err)
	}
	if _, err := db.Exec(
		"INSERT INTO model_responses (model_id, prompt_id, response_text) VALUES (?, ?, ?)",
		modelID, promptID, "saved response",
	); err != nil {
		t.Fatalf("failed to seed model response: %v", err)
	}

	form := url.Values{}
	form.Add("new_model_name", "NewModel")
	req := httptest.NewRequest("POST", "/edit_model?model=OldModel", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	EditModelHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d for successful rename, got %d", http.StatusSeeOther, rr.Code)
	}

	var responseText string
	err := db.QueryRow(`
		SELECT mr.response_text
		FROM model_responses mr
		JOIN models m ON mr.model_id = m.id
		WHERE m.name = 'NewModel' AND mr.prompt_id = ?
	`, promptID).Scan(&responseText)
	if err != nil {
		t.Fatalf("expected model response to survive the rename: %v", err)
	}
	if responseText != "saved response" {
		t.Fatalf("expected saved response text, got %q", responseText)
	}

	results := middleware.ReadResults()
	if scores := results["NewModel"].Scores; len(scores) != 1 || scores[0] != 42 {
		t.Fatalf("expected score 42 to survive the rename, got %v", scores)
	}
}

func TestEditModel_RenameModelError(t *testing.T) {
	mockDS := &MockDataStore{
		Results: map[string]middleware.Result{
			"OldModel": {Scores: []int{80}},
		},
	}
	mockDS.RenameModelFunc = func(suiteName, oldName, newName string) error {
		return errors.New("rename failed")
	}

	handler := NewHandlerWithDeps(mockDS, &MockRenderer{})

	form := url.Values{}
	form.Add("new_model_name", "NewModel")

	req := httptest.NewRequest("POST", "/edit_model?model=OldModel", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.EditModel(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for rename error, got %d", http.StatusInternalServerError, rr.Code)
	}
	if _, exists := mockDS.Results["OldModel"]; !exists {
		t.Error("results map must be untouched when the rename fails")
	}
}
