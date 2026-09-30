package testutil

import (
	"database/sql"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestSetupTestDB_CreatesSchemaAndAllowsInserts(t *testing.T) {
	db := SetupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })

	defaultSuiteID := GetDefaultSuiteID(t, db)
	if defaultSuiteID <= 0 {
		t.Fatalf("expected default suite id > 0, got %d", defaultSuiteID)
	}

	suiteID := CreateTestSuite(t, db, "suite-1")
	SetCurrentSuite(t, db, suiteID)

	var isCurrent int
	if err := db.QueryRow("SELECT is_current FROM suites WHERE id = ?", suiteID).Scan(&isCurrent); err != nil {
		t.Fatalf("failed to read suite current flag: %v", err)
	}
	if isCurrent != 1 {
		t.Fatalf("expected suite to be current, got is_current=%d", isCurrent)
	}

	profileID := CreateTestProfile(t, db, suiteID, "profile-1", "desc")

	promptWithProfileID := CreateTestPrompt(t, db, suiteID, "prompt-1", "sol-1", &profileID, 1, "objective")
	promptWithoutProfileID := CreateTestPrompt(t, db, suiteID, "prompt-2", "sol-2", nil, 2, "subjective")

	var withProfile sql.NullInt64
	if err := db.QueryRow("SELECT profile_id FROM prompts WHERE id = ?", promptWithProfileID).Scan(&withProfile); err != nil {
		t.Fatalf("failed to read prompt profile_id: %v", err)
	}
	if !withProfile.Valid || int(withProfile.Int64) != profileID {
		t.Fatalf("expected prompt profile_id to be %d, got %+v", profileID, withProfile)
	}

	var withoutProfile sql.NullInt64
	if err := db.QueryRow("SELECT profile_id FROM prompts WHERE id = ?", promptWithoutProfileID).Scan(&withoutProfile); err != nil {
		t.Fatalf("failed to read prompt profile_id: %v", err)
	}
	if withoutProfile.Valid {
		t.Fatalf("expected prompt profile_id to be NULL, got %+v", withoutProfile)
	}

	modelID := CreateTestModel(t, db, suiteID, "model-1")
	CreateTestScore(t, db, modelID, promptWithoutProfileID, 5)

	var score int
	if err := db.QueryRow("SELECT score FROM scores WHERE model_id = ? AND prompt_id = ?", modelID, promptWithoutProfileID).Scan(&score); err != nil {
		t.Fatalf("failed to read score: %v", err)
	}
	if score != 5 {
		t.Fatalf("expected score=5, got %d", score)
	}

	CreateTestModelResponse(t, db, modelID, promptWithoutProfileID, "hello")
	var response string
	if err := db.QueryRow("SELECT response_text FROM model_responses WHERE model_id = ? AND prompt_id = ?", modelID, promptWithoutProfileID).Scan(&response); err != nil {
		t.Fatalf("failed to read model response: %v", err)
	}
	if response != "hello" {
		t.Fatalf("expected response %q, got %q", "hello", response)
	}
}

func TestMockRenderer_Render_RecordsCallAndWritesContent(t *testing.T) {
	rec := httptest.NewRecorder()

	renderer := &MockRenderer{}
	if err := renderer.Render(rec, "example.tmpl", nil, map[string]string{"k": "v"}, "a.tmpl", "b.tmpl"); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}
	call := renderer.RenderCalls[0]
	if call.Name != "example.tmpl" {
		t.Fatalf("expected call.Name=%q, got %q", "example.tmpl", call.Name)
	}
	if got := rec.Body.String(); got != "mock rendered" {
		t.Fatalf("expected response body %q, got %q", "mock rendered", got)
	}
}

func TestMockRenderer_Render_ReturnsErrorWithoutWriting(t *testing.T) {
	rec := httptest.NewRecorder()

	renderer := &MockRenderer{RenderError: errors.New("boom")}
	if err := renderer.Render(rec, "example.tmpl", nil, nil, "a.tmpl"); err == nil {
		t.Fatalf("expected Render to return an error")
	}

	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}
	if got := rec.Body.String(); got != "" {
		t.Fatalf("expected response body to be empty on error, got %q", got)
	}
}

func TestMockRenderer_RenderTemplateSimple_DelegatesToRender(t *testing.T) {
	rec := httptest.NewRecorder()

	renderer := &MockRenderer{}
	if err := renderer.RenderTemplateSimple(rec, "example.tmpl", 123); err != nil {
		t.Fatalf("RenderTemplateSimple returned error: %v", err)
	}

	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}
	call := renderer.RenderCalls[0]
	if call.Name != "example.tmpl" {
		t.Fatalf("expected call.Name=%q, got %q", "example.tmpl", call.Name)
	}
	if len(call.Files) != 1 || call.Files[0] != "templates/example.tmpl" {
		t.Fatalf("expected call.Files to include templates path, got %#v", call.Files)
	}
}

func TestMockDataStore_DefaultBehaviorAndState(t *testing.T) {
	mock := &MockDataStore{}

	if got, err := mock.GetCurrentSuiteID(); err != nil || got != 1 {
		t.Fatalf("GetCurrentSuiteID() expected (1, nil), got (%d, %v)", got, err)
	}
	if got := mock.GetCurrentSuiteName(); got != "default" {
		t.Fatalf("GetCurrentSuiteName() expected %q, got %q", "default", got)
	}
	if got, err := mock.ListSuites(); err != nil || len(got) != 1 || got[0] != "default" {
		t.Fatalf("ListSuites() expected ([default], nil), got (%v, %v)", got, err)
	}
	if err := mock.SetCurrentSuite("suite-1"); err != nil {
		t.Fatalf("SetCurrentSuite returned error: %v", err)
	}
	if got := mock.GetCurrentSuiteName(); got != "suite-1" {
		t.Fatalf("GetCurrentSuiteName() expected updated value %q, got %q", "suite-1", got)
	}
	if !mock.SuiteExists("anything") {
		t.Fatalf("SuiteExists() expected true")
	}

	if got := mock.ReadPrompts(); got != nil {
		t.Fatalf("ReadPrompts() expected nil, got %#v", got)
	}
	prompts := []Prompt{{Text: "p1"}}
	if err := mock.WritePrompts(prompts); err != nil {
		t.Fatalf("WritePrompts returned error: %v", err)
	}
	if got := mock.ReadPrompts(); len(got) != 1 || got[0].Text != "p1" {
		t.Fatalf("ReadPrompts() expected written prompts, got %#v", got)
	}

	if got, err := mock.ReadPromptSuite("suite-1"); err != nil || len(got) != 1 || got[0].Text != "p1" {
		t.Fatalf("ReadPromptSuite() expected written prompts, got (%#v, %v)", got, err)
	}
	if err := mock.WritePromptSuite("suite-1", []Prompt{{Text: "p2"}}); err != nil {
		t.Fatalf("WritePromptSuite returned error: %v", err)
	}
	if got := mock.ReadPrompts(); len(got) != 1 || got[0].Text != "p2" {
		t.Fatalf("ReadPrompts() expected updated prompts, got %#v", got)
	}

	if got, err := mock.ListPromptSuites(); err != nil || len(got) != 1 || got[0] != "default" {
		t.Fatalf("ListPromptSuites() expected ([default], nil), got (%v, %v)", got, err)
	}

	mock.UpdatePromptsOrder([]int{1, 2, 3})

	if got := mock.ReadProfiles(); got != nil {
		t.Fatalf("ReadProfiles() expected nil, got %#v", got)
	}
	profiles := []Profile{{Name: "prof"}}
	if err := mock.WriteProfiles(profiles); err != nil {
		t.Fatalf("WriteProfiles returned error: %v", err)
	}
	if got := mock.ReadProfiles(); len(got) != 1 || got[0].Name != "prof" {
		t.Fatalf("ReadProfiles() expected written profiles, got %#v", got)
	}

	if got := mock.ReadResults(); got == nil || len(got) != 0 {
		t.Fatalf("ReadResults() expected empty non-nil map, got %#v", got)
	}
	results := map[string]Result{"model": {Scores: []int{1, 2, 3}}}
	if err := mock.WriteResults("suite-1", results); err != nil {
		t.Fatalf("WriteResults returned error: %v", err)
	}
	if got := mock.ReadResults(); len(got) != 1 || len(got["model"].Scores) != 3 {
		t.Fatalf("ReadResults() expected written results, got %#v", got)
	}

	mock.BroadcastResults()
}

func TestMockDataStore_ErrorsWhenConfigured(t *testing.T) {
	expectedErr := errors.New("expected error")
	mock := &MockDataStore{Err: expectedErr}

	if got, err := mock.GetCurrentSuiteID(); err != expectedErr || got != 0 {
		t.Fatalf("GetCurrentSuiteID() expected (0, %v), got (%d, %v)", expectedErr, got, err)
	}
	if got, err := mock.ListSuites(); err != expectedErr || got != nil {
		t.Fatalf("ListSuites() expected (nil, %v), got (%v, %v)", expectedErr, got, err)
	}
	if err := mock.SetCurrentSuite("suite-x"); err != expectedErr {
		t.Fatalf("SetCurrentSuite() expected error %v, got %v", expectedErr, err)
	}
	if mock.CurrentSuite != "" {
		t.Fatalf("expected CurrentSuite to remain empty on error, got %q", mock.CurrentSuite)
	}
	if err := mock.WritePrompts([]Prompt{{Text: "p"}}); err != expectedErr {
		t.Fatalf("WritePrompts() expected error %v, got %v", expectedErr, err)
	}
	if got, err := mock.ReadPromptSuite("suite-x"); err != expectedErr || got != nil {
		t.Fatalf("ReadPromptSuite() expected (nil, %v), got (%v, %v)", expectedErr, got, err)
	}
	if err := mock.WritePromptSuite("suite-x", []Prompt{{Text: "p"}}); err != expectedErr {
		t.Fatalf("WritePromptSuite() expected error %v, got %v", expectedErr, err)
	}
	if got, err := mock.ListPromptSuites(); err != expectedErr || got != nil {
		t.Fatalf("ListPromptSuites() expected (nil, %v), got (%v, %v)", expectedErr, got, err)
	}
	if err := mock.WriteProfiles([]Profile{{Name: "p"}}); err != expectedErr {
		t.Fatalf("WriteProfiles() expected error %v, got %v", expectedErr, err)
	}
	if got, err := mock.ReadPromptSuite("suite-x"); err != expectedErr || got != nil {
		t.Fatalf("ReadPromptSuite() expected (nil, %v), got (%v, %v)", expectedErr, got, err)
	}
	if err := mock.WriteResults("suite-x", map[string]Result{}); err != expectedErr {
		t.Fatalf("WriteResults() expected error %v, got %v", expectedErr, err)
	}
}

func TestMockDataStore_FunctionHooksTakePrecedence(t *testing.T) {
	var called struct {
		getCurrentSuiteID   bool
		getCurrentSuiteName bool
		listSuites          bool
		setCurrentSuite     bool
		suiteExists         bool
		readPrompts         bool
		writePrompts        bool
		readPromptSuite     bool
		writePromptSuite    bool
		listPromptSuites    bool
		updatePromptsOrder  bool
		readProfiles        bool
		writeProfiles       bool
		readResults         bool
		writeResults        bool
		broadcastResults    bool
	}

	mock := &MockDataStore{
		GetCurrentSuiteIDFunc: func() (int, error) {
			called.getCurrentSuiteID = true
			return 42, nil
		},
		GetCurrentSuiteNameFunc: func() string {
			called.getCurrentSuiteName = true
			return "suite-from-hook"
		},
		ListSuitesFunc: func() ([]string, error) {
			called.listSuites = true
			return []string{"a", "b"}, nil
		},
		SetCurrentSuiteFunc: func(name string) error {
			called.setCurrentSuite = true
			if name != "suite-z" {
				t.Fatalf("SetCurrentSuiteFunc called with %q", name)
			}
			return nil
		},
		SuiteExistsFunc: func(name string) bool {
			called.suiteExists = true
			return name == "exists"
		},
		ReadPromptsFunc: func() []Prompt {
			called.readPrompts = true
			return []Prompt{{Text: "from-hook"}}
		},
		WritePromptsFunc: func(prompts []Prompt) error {
			called.writePrompts = true
			if len(prompts) != 1 || prompts[0].Text != "p" {
				t.Fatalf("WritePromptsFunc called with %#v", prompts)
			}
			return nil
		},
		ReadPromptSuiteFunc: func(suiteName string) ([]Prompt, error) {
			called.readPromptSuite = true
			if suiteName != "suite-x" {
				t.Fatalf("ReadPromptSuiteFunc called with %q", suiteName)
			}
			return []Prompt{{Text: "suite-prompt"}}, nil
		},
		WritePromptSuiteFunc: func(suiteName string, prompts []Prompt) error {
			called.writePromptSuite = true
			if suiteName != "suite-x" {
				t.Fatalf("WritePromptSuiteFunc called with %q", suiteName)
			}
			if len(prompts) != 1 || prompts[0].Text != "p2" {
				t.Fatalf("WritePromptSuiteFunc called with %#v", prompts)
			}
			return nil
		},
		ListPromptSuitesFunc: func() ([]string, error) {
			called.listPromptSuites = true
			return []string{"suite-x"}, nil
		},
		UpdatePromptsOrderFunc: func(order []int) {
			called.updatePromptsOrder = true
			if len(order) != 3 || order[0] != 1 || order[2] != 3 {
				t.Fatalf("UpdatePromptsOrderFunc called with %#v", order)
			}
		},
		ReadProfilesFunc: func() []Profile {
			called.readProfiles = true
			return []Profile{{Name: "profile-from-hook"}}
		},
		WriteProfilesFunc: func(profiles []Profile) error {
			called.writeProfiles = true
			if len(profiles) != 1 || profiles[0].Name != "p" {
				t.Fatalf("WriteProfilesFunc called with %#v", profiles)
			}
			return nil
		},
		ReadResultsFunc: func() map[string]Result {
			called.readResults = true
			return map[string]Result{"model": {Scores: []int{1}}}
		},
		WriteResultsFunc: func(suiteName string, results map[string]Result) error {
			called.writeResults = true
			if suiteName != "suite-x" {
				t.Fatalf("WriteResultsFunc called with %q", suiteName)
			}
			if len(results) != 1 {
				t.Fatalf("WriteResultsFunc called with %#v", results)
			}
			return nil
		},
		BroadcastResultsFunc: func() {
			called.broadcastResults = true
		},
	}

	if got, err := mock.GetCurrentSuiteID(); err != nil || got != 42 || !called.getCurrentSuiteID {
		t.Fatalf("GetCurrentSuiteID hook not applied: got (%d, %v), called=%v", got, err, called.getCurrentSuiteID)
	}
	if got := mock.GetCurrentSuiteName(); got != "suite-from-hook" || !called.getCurrentSuiteName {
		t.Fatalf("GetCurrentSuiteName hook not applied: got %q, called=%v", got, called.getCurrentSuiteName)
	}
	if got, err := mock.ListSuites(); err != nil || len(got) != 2 || !called.listSuites {
		t.Fatalf("ListSuites hook not applied: got (%v, %v), called=%v", got, err, called.listSuites)
	}
	if err := mock.SetCurrentSuite("suite-z"); err != nil || !called.setCurrentSuite {
		t.Fatalf("SetCurrentSuite hook not applied: err=%v, called=%v", err, called.setCurrentSuite)
	}
	if got := mock.SuiteExists("exists"); !got || !called.suiteExists {
		t.Fatalf("SuiteExists hook not applied: got=%v, called=%v", got, called.suiteExists)
	}
	if got := mock.ReadPrompts(); len(got) != 1 || got[0].Text != "from-hook" || !called.readPrompts {
		t.Fatalf("ReadPrompts hook not applied: got=%#v, called=%v", got, called.readPrompts)
	}
	if err := mock.WritePrompts([]Prompt{{Text: "p"}}); err != nil || !called.writePrompts {
		t.Fatalf("WritePrompts hook not applied: err=%v, called=%v", err, called.writePrompts)
	}
	if got, err := mock.ReadPromptSuite("suite-x"); err != nil || len(got) != 1 || got[0].Text != "suite-prompt" || !called.readPromptSuite {
		t.Fatalf("ReadPromptSuite hook not applied: got (%#v, %v), called=%v", got, err, called.readPromptSuite)
	}
	if err := mock.WritePromptSuite("suite-x", []Prompt{{Text: "p2"}}); err != nil || !called.writePromptSuite {
		t.Fatalf("WritePromptSuite hook not applied: err=%v, called=%v", err, called.writePromptSuite)
	}
	if got, err := mock.ListPromptSuites(); err != nil || len(got) != 1 || got[0] != "suite-x" || !called.listPromptSuites {
		t.Fatalf("ListPromptSuites hook not applied: got (%v, %v), called=%v", got, err, called.listPromptSuites)
	}
	mock.UpdatePromptsOrder([]int{1, 2, 3})
	if !called.updatePromptsOrder {
		t.Fatalf("UpdatePromptsOrder hook not applied")
	}
	if got := mock.ReadProfiles(); len(got) != 1 || got[0].Name != "profile-from-hook" || !called.readProfiles {
		t.Fatalf("ReadProfiles hook not applied: got=%#v, called=%v", got, called.readProfiles)
	}
	if err := mock.WriteProfiles([]Profile{{Name: "p"}}); err != nil || !called.writeProfiles {
		t.Fatalf("WriteProfiles hook not applied: err=%v, called=%v", err, called.writeProfiles)
	}
	if got := mock.ReadResults(); got["model"].Scores[0] != 1 || !called.readResults {
		t.Fatalf("ReadResults hook not applied: got=%#v, called=%v", got, called.readResults)
	}
	if err := mock.WriteResults("suite-x", map[string]Result{"model": {Scores: []int{1}}}); err != nil || !called.writeResults {
		t.Fatalf("WriteResults hook not applied: err=%v, called=%v", err, called.writeResults)
	}
	mock.BroadcastResults()
	if !called.broadcastResults {
		t.Fatalf("BroadcastResults hook not applied")
	}
}
