package middleware

import (
	"testing"
)

// Scores and model_responses cascade on prompt_id, so the suite writer must
// update surviving prompt rows in place. Rewriting the suite by delete and
// reinsert would hand every prompt a fresh AUTOINCREMENT id and silently wipe
// all scores and saved responses. Handlers round-trip prompts read from
// ReadPrompts back through WritePrompts, which is the flow mirrored here.

func TestWritePromptSuite_PreservesScoresAcrossRewrite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WritePromptSuite("default", []Prompt{
		{Text: "p1", Solution: "s1"},
		{Text: "p2", Solution: "s2"},
	}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
	if err := WriteResults("default", map[string]Result{
		"ModelA": {Scores: []int{7, 9}},
	}); err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	prompts := ReadPrompts()
	prompts[1].Solution = "edited"
	if err := WritePrompts(prompts); err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	results := ReadResults()
	got := results["ModelA"].Scores
	if len(got) != 2 || got[0] != 7 || got[1] != 9 {
		t.Fatalf("expected scores [7 9] to survive the rewrite, got %v", got)
	}

	readBack, err := ReadPromptSuite("default")
	if err != nil {
		t.Fatalf("ReadPromptSuite failed: %v", err)
	}
	if readBack[1].Solution != "edited" {
		t.Fatalf("expected edit to persist, got solution %q", readBack[1].Solution)
	}
}

func TestWritePromptSuite_PreservesModelResponsesAcrossRewrite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
	if err := WriteResults("default", map[string]Result{
		"ModelA": {Scores: []int{42}},
	}); err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}
	modelID, promptID := modelAndPromptIDs(t, "ModelA", "p1")
	if _, err := db.Exec(
		"INSERT INTO model_responses (model_id, prompt_id, response_text) VALUES (?, ?, ?)",
		modelID, promptID, "saved response",
	); err != nil {
		t.Fatalf("failed to seed model response: %v", err)
	}

	prompts := ReadPrompts()
	prompts[0].Solution = "edited"
	if err := WritePrompts(prompts); err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	var responseText string
	newPromptID := promptIDByText(t, "p1")
	err := db.QueryRow(
		"SELECT response_text FROM model_responses WHERE model_id = ? AND prompt_id = ?",
		modelID, newPromptID,
	).Scan(&responseText)
	if err != nil {
		t.Fatalf("expected model response to survive the rewrite: %v", err)
	}
	if responseText != "saved response" {
		t.Fatalf("expected saved response text, got %q", responseText)
	}
}

func TestWritePromptSuite_DeleteOnlyCascadesRemovedPrompt(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WritePromptSuite("default", []Prompt{
		{Text: "p1"},
		{Text: "p2"},
	}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
	if err := WriteResults("default", map[string]Result{
		"ModelA": {Scores: []int{7, 9}},
	}); err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}
	modelID, _ := modelAndPromptIDs(t, "ModelA", "p1")
	p2ID := promptIDByText(t, "p2")
	for _, promptID := range []int{promptIDByText(t, "p1"), p2ID} {
		if _, err := db.Exec(
			"INSERT INTO model_responses (model_id, prompt_id, response_text) VALUES (?, ?, ?)",
			modelID, promptID, "resp",
		); err != nil {
			t.Fatalf("failed to seed model response: %v", err)
		}
	}

	// Drop p2 from the list, the way DeletePrompt does.
	prompts := ReadPrompts()
	prompts = prompts[:1]
	if err := WritePrompts(prompts); err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	results := ReadResults()
	got := results["ModelA"].Scores
	if len(got) != 1 || got[0] != 7 {
		t.Fatalf("expected surviving score [7], got %v", got)
	}

	var keptResponses, removedResponses int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM model_responses WHERE model_id = ? AND prompt_id = ?",
		modelID, promptIDByText(t, "p1"),
	).Scan(&keptResponses); err != nil {
		t.Fatalf("failed to count kept responses: %v", err)
	}
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM model_responses WHERE model_id = ? AND prompt_id = ?",
		modelID, p2ID,
	).Scan(&removedResponses); err != nil {
		t.Fatalf("failed to count removed responses: %v", err)
	}
	if keptResponses != 1 {
		t.Fatalf("expected response for kept prompt to survive, got %d", keptResponses)
	}
	if removedResponses != 0 {
		t.Fatalf("expected response for removed prompt to cascade away, got %d", removedResponses)
	}
}

func TestWritePromptSuite_RewriteUpdatesDisplayOrder(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WritePromptSuite("default", []Prompt{
		{Text: "p1"},
		{Text: "p2"},
	}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
	if err := WriteResults("default", map[string]Result{
		"ModelA": {Scores: []int{7, 9}},
	}); err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	// Reverse the list, the way MovePrompt reorders entries.
	prompts := ReadPrompts()
	prompts = []Prompt{prompts[1], prompts[0]}
	if err := WritePrompts(prompts); err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	results := ReadResults()
	got := results["ModelA"].Scores
	if len(got) != 2 || got[0] != 9 || got[1] != 7 {
		t.Fatalf("expected scores to follow the move as [9 7], got %v", got)
	}
}

func TestWritePromptSuite_StaleIDInsertsFreshRow(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WritePromptSuite("default", []Prompt{{Text: "p1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	prompts := ReadPrompts()
	prompts[0].ID = 1 << 30
	if err := WritePrompts(prompts); err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}

	var rowID int
	if err := db.QueryRow("SELECT id FROM prompts WHERE text = 'p1'").Scan(&rowID); err != nil {
		t.Fatalf("failed to read prompt row: %v", err)
	}
	if rowID == 1<<30 {
		t.Fatalf("expected a fresh row id for a stale incoming ID, got %d", rowID)
	}
}

func TestWritePromptSuite_KeepsProfileResolutionAcrossRewrite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WriteProfileSuite("default", []Profile{{Name: "P1", Description: "d"}}); err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{{Text: "p1", Profile: "P1"}}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}

	prompts := ReadPrompts()
	if err := WritePrompts(prompts); err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}
	readBack, _ := ReadPromptSuite("default")
	if readBack[0].Profile != "P1" {
		t.Fatalf("expected profile link to survive the rewrite, got %q", readBack[0].Profile)
	}

	// A name that no longer resolves must read back unprofiled, not fail.
	prompts = ReadPrompts()
	prompts[0].Profile = "ghost"
	if err := WritePrompts(prompts); err != nil {
		t.Fatalf("WritePrompts failed: %v", err)
	}
	readBack, _ = ReadPromptSuite("default")
	if readBack[0].Profile != "" {
		t.Fatalf("expected unresolvable profile to read back empty, got %q", readBack[0].Profile)
	}
}

func modelAndPromptIDs(t *testing.T, modelName, promptText string) (int, int) {
	t.Helper()

	var modelID int
	if err := db.QueryRow("SELECT id FROM models WHERE name = ?", modelName).Scan(&modelID); err != nil {
		t.Fatalf("failed to read model id: %v", err)
	}
	return modelID, promptIDByText(t, promptText)
}

func promptIDByText(t *testing.T, promptText string) int {
	t.Helper()

	var promptID int
	if err := db.QueryRow("SELECT id FROM prompts WHERE text = ?", promptText).Scan(&promptID); err != nil {
		t.Fatalf("failed to read prompt id for %q: %v", promptText, err)
	}
	return promptID
}
