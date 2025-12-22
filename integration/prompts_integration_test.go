package integration

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"llm-tournament/handlers"
	"llm-tournament/middleware"

	_ "github.com/mattn/go-sqlite3"
)

// setupIntegrationDB creates a test database for integration tests
func setupIntegrationDB(t *testing.T) func() {
	t.Helper()
	dbPath := t.TempDir() + "/integration_test.db"
	err := middleware.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize integration test database: %v", err)
	}
	return func() {
		_ = middleware.CloseDB()
	}
}

// TestPromptManagementWorkflow tests the complete prompt management lifecycle
func TestPromptManagementWorkflow(t *testing.T) {
	cleanup := setupIntegrationDB(t)
	defer cleanup()

	// Step 1: Create a new suite
	form := url.Values{}
	form.Add("suite_name", "Integration Test Suite")

	req := httptest.NewRequest("POST", "/prompts/suites/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handlers.NewPromptSuiteHandler(rr, req)

	if rr.Code != 303 {
		t.Fatalf("Step 1 failed: expected status 303, got %d", rr.Code)
	}

	// Verify suite was created (plus default suite)
	suites, err := middleware.ListSuites()
	if err != nil {
		t.Fatalf("Step 1 failed: error listing suites: %v", err)
	}
	found := false
	for _, suite := range suites {
		if suite == "Integration Test Suite" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Step 1 failed: 'Integration Test Suite' not found in suites list")
	}

	// Step 2: Add profiles
	profiles := []string{"Math", "Science", "History"}
	for _, profileName := range profiles {
		form := url.Values{}
		form.Add("profile_name", profileName)

		req := httptest.NewRequest("POST", "/add_profile", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handlers.AddProfileHandler(rr, req)

		if rr.Code != 303 {
			t.Fatalf("Step 2 failed: expected status 303 for profile %q, got %d", profileName, rr.Code)
		}
	}

	// Verify profiles were created
	createdProfiles := middleware.ReadProfiles()
	if len(createdProfiles) != 3 {
		t.Fatalf("Step 2 failed: expected 3 profiles, got %d", len(createdProfiles))
	}

	// Step 3: Add prompts to different profiles
	prompts := []struct {
		text     string
		solution string
		profile  string
	}{
		{"What is 2+2?", "4", "Math"},
		{"What is photosynthesis?", "Process by which plants make food", "Science"},
		{"When did WW2 end?", "1945", "History"},
		{"What is 5*6?", "30", "Math"},
		{"What is gravity?", "Force of attraction", "Science"},
	}

	for _, p := range prompts {
		form := url.Values{}
		form.Add("prompt", p.text)
		form.Add("solution", p.solution)
		form.Add("profile", p.profile)

		req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handlers.AddPromptHandler(rr, req)

		if rr.Code != 303 {
			t.Fatalf("Step 3 failed: expected status 303 for prompt %q, got %d", p.text, rr.Code)
		}
	}

	// Verify prompts were added
	createdPrompts := middleware.ReadPrompts()
	if len(createdPrompts) != 5 {
		t.Fatalf("Step 3 failed: expected 5 prompts, got %d", len(createdPrompts))
	}

	// Step 4: Verify prompts are returned in correct order
	// ReadPrompts() orders by display_order internally
	// Just verify the prompts are in the expected sequence
	expectedFirstPrompt := "What is 2+2?"
	if createdPrompts[0].Text != expectedFirstPrompt {
		t.Errorf("Step 4 failed: expected first prompt %q, got %q", expectedFirstPrompt, createdPrompts[0].Text)
	}

	// Verify profile grouping
	mathCount := 0
	scienceCount := 0
	historyCount := 0
	for _, prompt := range createdPrompts {
		switch prompt.Profile {
		case "Math":
			mathCount++
		case "Science":
			scienceCount++
		case "History":
			historyCount++
		}
	}
	if mathCount != 2 {
		t.Errorf("Step 4 failed: expected 2 Math prompts, got %d", mathCount)
	}
	if scienceCount != 2 {
		t.Errorf("Step 4 failed: expected 2 Science prompts, got %d", scienceCount)
	}
	if historyCount != 1 {
		t.Errorf("Step 4 failed: expected 1 History prompt, got %d", historyCount)
	}

	// Step 5: Delete a profile and verify cascade delete
	// Find index of "History" profile
	historyIndex := -1
	for i, profile := range createdProfiles {
		if profile.Name == "History" {
			historyIndex = i
			break
		}
	}
	if historyIndex == -1 {
		t.Fatal("Step 5 failed: History profile not found")
	}

	form = url.Values{}
	form.Add("index", fmt.Sprintf("%d", historyIndex))

	req = httptest.NewRequest("POST", "/delete_profile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handlers.DeleteProfileHandler(rr, req)

	if rr.Code != 303 {
		t.Fatalf("Step 5 failed: expected status 303, got %d", rr.Code)
	}

	// Verify profile was deleted
	remainingProfiles := middleware.ReadProfiles()
	if len(remainingProfiles) != 2 {
		t.Fatalf("Step 5 failed: expected 2 profiles after deletion, got %d", len(remainingProfiles))
	}

	// Verify prompts associated with deleted profile are removed (or profile set to empty)
	remainingPrompts := middleware.ReadPrompts()
	historyPromptsCount := 0
	for _, prompt := range remainingPrompts {
		if prompt.Profile == "History" {
			historyPromptsCount++
		}
	}
	if historyPromptsCount > 0 {
		t.Errorf("Step 5 failed: expected 0 History prompts after profile deletion, got %d", historyPromptsCount)
	}

	// Step 6: Edit a prompt
	if len(remainingPrompts) > 0 {
		form := url.Values{}
		form.Add("index", "0")
		form.Add("prompt", "What is 2+2? (edited)")
		form.Add("solution", "4 (verified)")
		form.Add("profile", "Math")

		req := httptest.NewRequest("POST", "/edit_prompt", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handlers.EditPromptHandler(rr, req)

		if rr.Code != 303 {
			t.Fatalf("Step 6 failed: expected status 303, got %d", rr.Code)
		}

		// Verify edit
		editedPrompts := middleware.ReadPrompts()
		if len(editedPrompts) == 0 {
			t.Fatal("Step 6 failed: no prompts found after edit")
		}
		if !strings.Contains(editedPrompts[0].Text, "(edited)") {
			t.Errorf("Step 6 failed: expected edited prompt text, got %q", editedPrompts[0].Text)
		}
	}

	// Step 7: Delete suite and verify all related data is removed
	form = url.Values{}
	form.Add("suite_name", "Integration Test Suite")

	req = httptest.NewRequest("POST", "/prompts/suites/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	handlers.DeletePromptSuiteHandler(rr, req)

	if rr.Code != 303 {
		t.Fatalf("Step 7 failed: expected status 303, got %d", rr.Code)
	}

	// Verify suite was deleted
	remainingSuites, err := middleware.ListSuites()
	if err != nil {
		t.Fatalf("Step 7 failed: error listing suites: %v", err)
	}
	suiteFound := false
	for _, suite := range remainingSuites {
		if suite == "Integration Test Suite" {
			suiteFound = true
			break
		}
	}
	if suiteFound {
		t.Error("Step 7 failed: suite should be deleted")
	}
}

// TestDisplayOrderPreservation tests that display_order is correctly maintained
func TestDisplayOrderPreservation(t *testing.T) {
	cleanup := setupIntegrationDB(t)
	defer cleanup()

	// Add several prompts
	for i := 0; i < 5; i++ {
		form := url.Values{}
		form.Add("prompt", "Prompt "+string(rune('A'+i)))
		form.Add("solution", "Solution")
		form.Add("profile", "")

		req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handlers.AddPromptHandler(rr, req)

		if rr.Code != 303 {
			t.Fatalf("Failed to add prompt %d: status %d", i, rr.Code)
		}
	}

	// Verify prompts are returned in correct order (by display_order internally)
	prompts := middleware.ReadPrompts()
	if len(prompts) != 5 {
		t.Fatalf("expected 5 prompts, got %d", len(prompts))
	}

	// Verify order matches insertion order
	expectedOrder := []string{"Prompt A", "Prompt B", "Prompt C", "Prompt D", "Prompt E"}
	for i, prompt := range prompts {
		if prompt.Text != expectedOrder[i] {
			t.Errorf("expected prompts[%d].Text = %q, got %q", i, expectedOrder[i], prompt.Text)
		}
	}

	// Delete middle prompt
	form := url.Values{}
	form.Add("index", "2")

	req := httptest.NewRequest("POST", "/delete_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handlers.DeletePromptHandler(rr, req)

	if rr.Code != 303 {
		t.Fatalf("Failed to delete prompt: status %d", rr.Code)
	}

	// Verify remaining prompts are in correct order
	remainingPrompts := middleware.ReadPrompts()
	if len(remainingPrompts) != 4 {
		t.Fatalf("expected 4 prompts after deletion, got %d", len(remainingPrompts))
	}

	// Should have Prompt A, B, D, E (C was deleted)
	expectedAfterDelete := []string{"Prompt A", "Prompt B", "Prompt D", "Prompt E"}
	for i, prompt := range remainingPrompts {
		if prompt.Text != expectedAfterDelete[i] {
			t.Errorf("expected prompts[%d].Text = %q after deletion, got %q", i, expectedAfterDelete[i], prompt.Text)
		}
	}
}

// TestMultipleSuitesIsolation tests that suites are properly isolated
func TestMultipleSuitesIsolation(t *testing.T) {
	cleanup := setupIntegrationDB(t)
	defer cleanup()

	// Create first suite
	form1 := url.Values{}
	form1.Add("suite_name", "Suite A")
	req1 := httptest.NewRequest("POST", "/prompts/suites/new", strings.NewReader(form1.Encode()))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr1 := httptest.NewRecorder()
	handlers.NewPromptSuiteHandler(rr1, req1)

	// Select Suite A to make it current
	formSelectInitial := url.Values{}
	formSelectInitial.Add("suite_name", "Suite A")
	reqSelectInitial := httptest.NewRequest("POST", "/prompts/suites/select", strings.NewReader(formSelectInitial.Encode()))
	reqSelectInitial.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rrSelectInitial := httptest.NewRecorder()
	handlers.SelectPromptSuiteHandler(rrSelectInitial, reqSelectInitial)

	// Add prompts to Suite A
	form := url.Values{}
	form.Add("prompt", "Suite A Prompt")
	form.Add("solution", "Solution A")
	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handlers.AddPromptHandler(rr, req)

	promptsA := middleware.ReadPrompts()
	if len(promptsA) != 1 {
		t.Fatalf("Suite A should have 1 prompt, got %d", len(promptsA))
	}

	// Create second suite
	form2 := url.Values{}
	form2.Add("suite_name", "Suite B")
	req2 := httptest.NewRequest("POST", "/prompts/suites/new", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr2 := httptest.NewRecorder()
	handlers.NewPromptSuiteHandler(rr2, req2)

	// Select Suite B
	formSelect := url.Values{}
	formSelect.Add("suite_name", "Suite B")
	reqSelect := httptest.NewRequest("POST", "/prompts/suites/select", strings.NewReader(formSelect.Encode()))
	reqSelect.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rrSelect := httptest.NewRecorder()
	handlers.SelectPromptSuiteHandler(rrSelect, reqSelect)

	// Suite B should start empty
	promptsB := middleware.ReadPrompts()
	if len(promptsB) != 0 {
		t.Fatalf("Suite B should start with 0 prompts, got %d", len(promptsB))
	}

	// Add prompts to Suite B
	form3 := url.Values{}
	form3.Add("prompt", "Suite B Prompt")
	form3.Add("solution", "Solution B")
	req3 := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form3.Encode()))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr3 := httptest.NewRecorder()
	handlers.AddPromptHandler(rr3, req3)

	promptsB = middleware.ReadPrompts()
	if len(promptsB) != 1 {
		t.Fatalf("Suite B should now have 1 prompt, got %d", len(promptsB))
	}

	// Switch back to Suite A
	formSelectA := url.Values{}
	formSelectA.Add("suite_name", "Suite A")
	reqSelectA := httptest.NewRequest("POST", "/prompts/suites/select", strings.NewReader(formSelectA.Encode()))
	reqSelectA.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rrSelectA := httptest.NewRecorder()
	handlers.SelectPromptSuiteHandler(rrSelectA, reqSelectA)

	// Suite A should still have its original prompt
	promptsAAgain := middleware.ReadPrompts()
	if len(promptsAAgain) != 1 {
		t.Fatalf("Suite A should still have 1 prompt, got %d", len(promptsAAgain))
	}
	if promptsAAgain[0].Text != "Suite A Prompt" {
		t.Errorf("Suite A prompt should be preserved, got %q", promptsAAgain[0].Text)
	}
}
