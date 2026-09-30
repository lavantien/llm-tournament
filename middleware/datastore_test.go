package middleware

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestDefaultDataStore_IsSet(t *testing.T) {
	if DefaultDataStore == nil {
		t.Fatal("DefaultDataStore should not be nil")
	}
}

func TestDefaultDataStore_IsSQLiteDataStore(t *testing.T) {
	_, ok := DefaultDataStore.(*SQLiteDataStore)
	if !ok {
		t.Error("DefaultDataStore should be a *SQLiteDataStore")
	}
}

func TestDefaultDataStore_CanBeSwapped(t *testing.T) {
	// Save original
	original := DefaultDataStore
	defer func() { DefaultDataStore = original }()

	// Create mock
	mock := &MockDataStore{}
	DefaultDataStore = mock

	if DefaultDataStore != mock {
		t.Error("DefaultDataStore should be the mock")
	}
}

// MockDataStore implements DataStore for testing
type MockDataStore struct {
	GetCurrentSuiteIDFunc   func() (int, error)
	GetCurrentSuiteNameFunc func() string
	ListSuitesFunc          func() ([]string, error)
	SetCurrentSuiteFunc     func(name string) error
	SuiteExistsFunc         func(name string) bool
	ReadPromptsFunc         func() []Prompt
	WritePromptsFunc        func(prompts []Prompt) error
	ReadPromptSuiteFunc     func(suiteName string) ([]Prompt, error)
	WritePromptSuiteFunc    func(suiteName string, prompts []Prompt) error
	ListPromptSuitesFunc    func() ([]string, error)
	UpdatePromptsOrderFunc  func(order []int)
	ReadProfilesFunc        func() []Profile
	WriteProfilesFunc       func(profiles []Profile) error
	ReadResultsFunc         func() map[string]Result
	WriteResultsFunc        func(suiteName string, results map[string]Result) error
	RenameModelFunc         func(suiteName, oldName, newName string) error
	BroadcastResultsFunc    func()

	Err      error
	Prompts  []Prompt
	Profiles []Profile
	Results  map[string]Result
}

func (m *MockDataStore) GetCurrentSuiteID() (int, error) {
	if m.GetCurrentSuiteIDFunc != nil {
		return m.GetCurrentSuiteIDFunc()
	}
	if m.Err != nil {
		return 0, m.Err
	}
	return 1, nil
}

func (m *MockDataStore) GetCurrentSuiteName() string {
	if m.GetCurrentSuiteNameFunc != nil {
		return m.GetCurrentSuiteNameFunc()
	}
	return "default"
}

func (m *MockDataStore) ListSuites() ([]string, error) {
	if m.ListSuitesFunc != nil {
		return m.ListSuitesFunc()
	}
	if m.Err != nil {
		return nil, m.Err
	}
	return []string{"default"}, nil
}

func (m *MockDataStore) SetCurrentSuite(name string) error {
	if m.SetCurrentSuiteFunc != nil {
		return m.SetCurrentSuiteFunc(name)
	}
	return m.Err
}

func (m *MockDataStore) SuiteExists(name string) bool {
	if m.SuiteExistsFunc != nil {
		return m.SuiteExistsFunc(name)
	}
	return true
}

func (m *MockDataStore) ReadPrompts() []Prompt {
	if m.ReadPromptsFunc != nil {
		return m.ReadPromptsFunc()
	}
	return m.Prompts
}

func (m *MockDataStore) WritePrompts(prompts []Prompt) error {
	if m.WritePromptsFunc != nil {
		return m.WritePromptsFunc(prompts)
	}
	if m.Err != nil {
		return m.Err
	}
	m.Prompts = prompts
	return nil
}

func (m *MockDataStore) ReadPromptSuite(suiteName string) ([]Prompt, error) {
	if m.ReadPromptSuiteFunc != nil {
		return m.ReadPromptSuiteFunc(suiteName)
	}
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Prompts, nil
}

func (m *MockDataStore) WritePromptSuite(suiteName string, prompts []Prompt) error {
	if m.WritePromptSuiteFunc != nil {
		return m.WritePromptSuiteFunc(suiteName, prompts)
	}
	if m.Err != nil {
		return m.Err
	}
	m.Prompts = prompts
	return nil
}

func (m *MockDataStore) ListPromptSuites() ([]string, error) {
	if m.ListPromptSuitesFunc != nil {
		return m.ListPromptSuitesFunc()
	}
	if m.Err != nil {
		return nil, m.Err
	}
	return []string{"default"}, nil
}

func (m *MockDataStore) UpdatePromptsOrder(order []int) {
	if m.UpdatePromptsOrderFunc != nil {
		m.UpdatePromptsOrderFunc(order)
	}
}

func (m *MockDataStore) ReadProfiles() []Profile {
	if m.ReadProfilesFunc != nil {
		return m.ReadProfilesFunc()
	}
	return m.Profiles
}

func (m *MockDataStore) WriteProfiles(profiles []Profile) error {
	if m.WriteProfilesFunc != nil {
		return m.WriteProfilesFunc(profiles)
	}
	if m.Err != nil {
		return m.Err
	}
	m.Profiles = profiles
	return nil
}

func (m *MockDataStore) ReadResults() map[string]Result {
	if m.ReadResultsFunc != nil {
		return m.ReadResultsFunc()
	}
	if m.Results == nil {
		return make(map[string]Result)
	}
	return m.Results
}

func (m *MockDataStore) WriteResults(suiteName string, results map[string]Result) error {
	if m.WriteResultsFunc != nil {
		return m.WriteResultsFunc(suiteName, results)
	}
	if m.Err != nil {
		return m.Err
	}
	m.Results = results
	return nil
}

func (m *MockDataStore) RenameModel(suiteName, oldName, newName string) error {
	if m.RenameModelFunc != nil {
		return m.RenameModelFunc(suiteName, oldName, newName)
	}
	if m.Err != nil {
		return m.Err
	}
	return nil
}

func (m *MockDataStore) BroadcastResults() {
	if m.BroadcastResultsFunc != nil {
		m.BroadcastResultsFunc()
	}
}

func TestMockDataStore_ReturnsError(t *testing.T) {
	expectedErr := errors.New("mock error")
	mock := &MockDataStore{Err: expectedErr}

	_, err := mock.GetCurrentSuiteID()
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	_, err = mock.ListSuites()
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	_, err = mock.ReadPromptSuite("test")
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	err = mock.WritePrompts(nil)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestMockDataStore_DefaultValues(t *testing.T) {
	mock := &MockDataStore{}

	if mock.GetCurrentSuiteName() != "default" {
		t.Error("expected default suite name")
	}

	if !mock.SuiteExists("test") {
		t.Error("expected SuiteExists to return true by default")
	}

	results := mock.ReadResults()
	if results == nil {
		t.Error("expected non-nil results map")
	}
}

// Tests for SQLiteDataStore wrapper methods
func TestSQLiteDataStore_GetCurrentSuiteID(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}
	id, err := ds.GetCurrentSuiteID()
	if err != nil {
		t.Errorf("GetCurrentSuiteID failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive suite ID, got %d", id)
	}
}

func TestSQLiteDataStore_GetCurrentSuiteName(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}
	name := ds.GetCurrentSuiteName()
	if name == "" {
		t.Error("expected non-empty suite name")
	}
}

func TestSQLiteDataStore_ListSuites(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}
	suites, err := ds.ListSuites()
	if err != nil {
		t.Errorf("ListSuites failed: %v", err)
	}
	if len(suites) == 0 {
		t.Error("expected at least one suite")
	}
}

func TestSQLiteDataStore_SetCurrentSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}
	err = ds.SetCurrentSuite("test-suite")
	if err != nil {
		t.Errorf("SetCurrentSuite failed: %v", err)
	}
}

func TestSQLiteDataStore_SuiteExists(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}
	if !ds.SuiteExists("default") {
		t.Error("expected default suite to exist")
	}
	if ds.SuiteExists("nonexistent-suite-xyz") {
		t.Error("expected nonexistent suite to not exist")
	}
}

func TestSQLiteDataStore_ReadWritePrompts(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}

	// Initially empty
	prompts := ds.ReadPrompts()
	if len(prompts) != 0 {
		t.Errorf("expected 0 prompts, got %d", len(prompts))
	}

	// Write prompts
	err = ds.WritePrompts([]Prompt{{Text: "Test prompt", Solution: "Test solution"}})
	if err != nil {
		t.Errorf("WritePrompts failed: %v", err)
	}

	// Read back
	prompts = ds.ReadPrompts()
	if len(prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(prompts))
	}
}

func TestSQLiteDataStore_ReadWritePromptSuite(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}

	// Write to suite
	err = ds.WritePromptSuite("test-suite", []Prompt{{Text: "Suite prompt"}})
	if err != nil {
		t.Errorf("WritePromptSuite failed: %v", err)
	}

	// Read back
	prompts, err := ds.ReadPromptSuite("test-suite")
	if err != nil {
		t.Errorf("ReadPromptSuite failed: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(prompts))
	}
}

func TestSQLiteDataStore_ListPromptSuites(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}
	suites, err := ds.ListPromptSuites()
	if err != nil {
		t.Errorf("ListPromptSuites failed: %v", err)
	}
	if len(suites) == 0 {
		t.Error("expected at least one suite")
	}
}

func TestSQLiteDataStore_UpdatePromptsOrder(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}

	// Write some prompts first
	_ = ds.WritePrompts([]Prompt{{Text: "P1"}, {Text: "P2"}})

	// Update order through the DataStore wrapper and verify it applied.
	ds.UpdatePromptsOrder([]int{1, 0})

	prompts := ds.ReadPrompts()
	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompts))
	}
	expectedSequence := []string{"P2", "P1"}
	for i, want := range expectedSequence {
		if prompts[i].Text != want {
			t.Errorf("prompts[%d] = %q, want %q", i, prompts[i].Text, want)
		}
	}
}

func TestSQLiteDataStore_ReadWriteProfiles(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}

	// Initially empty
	profiles := ds.ReadProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}

	// Write profiles
	err = ds.WriteProfiles([]Profile{{Name: "Test", Description: "Test profile"}})
	if err != nil {
		t.Errorf("WriteProfiles failed: %v", err)
	}

	// Read back
	profiles = ds.ReadProfiles()
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
}

func TestSQLiteDataStore_ReadWriteResults(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}

	// Write prompts first (needed for results)
	_ = ds.WritePrompts([]Prompt{{Text: "P1"}})

	// Write results
	err = ds.WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{80}},
	})
	if err != nil {
		t.Errorf("WriteResults failed: %v", err)
	}

	// Read back
	results := ds.ReadResults()
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestSQLiteDataStore_BroadcastResults(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}
	_ = ds.WritePrompts([]Prompt{{Text: "P1"}})
	err = ds.WriteResults("default", map[string]Result{
		"Model1": {Scores: []int{80}},
	})
	if err != nil {
		t.Fatalf("WriteResults failed: %v", err)
	}

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

	go ds.BroadcastResults()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read broadcast: %v", err)
	}

	var payload struct {
		Type string `json:"type"`
		Data struct {
			Results   map[string]Result `json:"results"`
			Models    []string          `json:"models"`
			SuiteName string            `json:"suiteName"`
			Prompts   []string          `json:"prompts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg, &payload); err != nil {
		t.Fatalf("failed to unmarshal broadcast: %v", err)
	}

	if payload.Type != "results" {
		t.Errorf("expected type 'results', got %q", payload.Type)
	}
	if payload.Data.SuiteName != "default" {
		t.Errorf("expected suite 'default', got %q", payload.Data.SuiteName)
	}
	if len(payload.Data.Models) != 1 || payload.Data.Models[0] != "Model1" {
		t.Errorf("expected models [Model1], got %v", payload.Data.Models)
	}
	if _, exists := payload.Data.Results["Model1"]; !exists {
		t.Errorf("expected results entry for Model1, got %#v", payload.Data.Results)
	}
	if len(payload.Data.Prompts) != 1 || payload.Data.Prompts[0] != "P1" {
		t.Errorf("expected prompts [P1], got %v", payload.Data.Prompts)
	}
}
