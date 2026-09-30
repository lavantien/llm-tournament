package middleware

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Prompts whose profile was deleted from the database must fall back to the
// Uncategorized group everywhere prompts are grouped, and no grouping path
// may panic. The schema nulls the dangling prompts.profile_id (ON DELETE SET
// NULL), so the prompt reads back with an empty profile string.
func TestDeletedProfilePromptsFallIntoUncategorized(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if err := WriteProfileSuite("default", []Profile{{Name: "temp", Description: "Temporary"}}); err != nil {
		t.Fatalf("WriteProfileSuite failed: %v", err)
	}
	if err := WritePromptSuite("default", []Prompt{
		{Text: "Profiled 1", Profile: "temp"},
		{Text: "Profiled 2", Profile: "temp"},
		{Text: "Bare"},
	}); err != nil {
		t.Fatalf("WritePromptSuite failed: %v", err)
	}
	if err := WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{50, 100, 80}},
	}); err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

	if _, err := db.Exec("DELETE FROM profiles WHERE name = 'temp'"); err != nil {
		t.Fatalf("failed to delete profile: %v", err)
	}

	// The dangling references must read back as unprofiled prompts.
	prompts := ReadPrompts()
	if len(prompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(prompts))
	}
	for _, prompt := range prompts {
		if prompt.Profile != "" {
			t.Fatalf("expected deleted profile to read back empty, got %q", prompt.Profile)
		}
	}

	// The shared grouping helper must not panic and yields no profile
	// groups for a fully unprofiled prompt set.
	groups, profileMap := GetProfileGroups(prompts, ReadProfiles())
	if len(groups) != 0 {
		t.Errorf("expected no profile groups after profile deletion, got %#v", groups)
	}
	if len(profileMap) != 0 {
		t.Errorf("expected empty profile map after profile deletion, got %#v", profileMap)
	}

	// The broadcast grouping path assigns the prompts to Uncategorized.
	clientsMutex.Lock()
	clients = make(map[*websocket.Conn]bool)
	clientsMutex.Unlock()

	server, wsURL := createWebSocketTestServer(t, HandleWebSocket)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	waitForWebSocketClientRegistration(t, 1)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	go BroadcastResults()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read broadcast: %v", err)
	}

	var payload struct {
		Type string `json:"type"`
		Data struct {
			ProfileGroups  []ProfileGroup `json:"profileGroups"`
			OrderedPrompts []struct {
				ProfileID   string `json:"profileId"`
				ProfileName string `json:"profileName"`
			} `json:"orderedPrompts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg, &payload); err != nil {
		t.Fatalf("failed to unmarshal broadcast: %v", err)
	}
	if payload.Type != "results" {
		t.Fatalf("expected type 'results', got %q", payload.Type)
	}

	if len(payload.Data.ProfileGroups) != 1 {
		t.Fatalf("expected only the Uncategorized group, got %#v", payload.Data.ProfileGroups)
	}
	uncategorized := payload.Data.ProfileGroups[0]
	if uncategorized.Name != "Uncategorized" || uncategorized.ID != "none" {
		t.Fatalf("expected Uncategorized/none group, got %q/%q", uncategorized.Name, uncategorized.ID)
	}
	if uncategorized.StartCol != 0 || uncategorized.EndCol != 2 {
		t.Fatalf("expected Uncategorized to span columns 0..2, got %d..%d", uncategorized.StartCol, uncategorized.EndCol)
	}

	if len(payload.Data.OrderedPrompts) != 3 {
		t.Fatalf("expected 3 ordered prompts, got %d", len(payload.Data.OrderedPrompts))
	}
	for i, op := range payload.Data.OrderedPrompts {
		if op.ProfileID != "none" || op.ProfileName != "" {
			t.Errorf("orderedPrompts[%d] = %q/%q, want none/\"\"", i, op.ProfileID, op.ProfileName)
		}
	}
}

// GetProfileGroups builds its map from the profile strings on the prompts,
// so a prompt carrying a profile name that is not in the profiles list still
// resolves to a real group instead of a nil pointer.
func TestGetProfileGroups_MissingProfileReferenceDoesNotPanic(t *testing.T) {
	prompts := []Prompt{
		{Text: "p1", Profile: "Ghost"},
		{Text: "p2", Profile: "Ghost"},
	}

	groups, profileMap := GetProfileGroups(prompts, []Profile{{Name: "Other"}})

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups (ghost + unused Other), got %d", len(groups))
	}
	if groups[0].Name != "Ghost" {
		t.Errorf("expected ghost profile group first, got %q", groups[0].Name)
	}
	ghost, exists := profileMap["Ghost"]
	if !exists || ghost == nil {
		t.Fatal("expected a non-nil group for the ghost profile reference")
	}
}
