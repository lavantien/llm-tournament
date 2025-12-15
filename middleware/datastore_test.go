package middleware

import (
	"errors"
	"testing"
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
	GetSettingFunc          func(key string) (string, error)
	SetSettingFunc          func(key, value string) error
	GetAPIKeyFunc           func(provider string) (string, error)
	SetAPIKeyFunc           func(provider, key string) error
	GetMaskedAPIKeysFunc    func() (map[string]string, error)
	BroadcastResultsFunc    func()

	Err      error
	Prompts  []Prompt
	Profiles []Profile
	Results  map[string]Result
	Settings map[string]string
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

func (m *MockDataStore) GetSetting(key string) (string, error) {
	if m.GetSettingFunc != nil {
		return m.GetSettingFunc(key)
	}
	if m.Err != nil {
		return "", m.Err
	}
	if m.Settings != nil {
		return m.Settings[key], nil
	}
	return "", nil
}

func (m *MockDataStore) SetSetting(key, value string) error {
	if m.SetSettingFunc != nil {
		return m.SetSettingFunc(key, value)
	}
	if m.Err != nil {
		return m.Err
	}
	if m.Settings == nil {
		m.Settings = make(map[string]string)
	}
	m.Settings[key] = value
	return nil
}

func (m *MockDataStore) GetAPIKey(provider string) (string, error) {
	if m.GetAPIKeyFunc != nil {
		return m.GetAPIKeyFunc(provider)
	}
	if m.Err != nil {
		return "", m.Err
	}
	return "", nil
}

func (m *MockDataStore) SetAPIKey(provider, key string) error {
	if m.SetAPIKeyFunc != nil {
		return m.SetAPIKeyFunc(provider, key)
	}
	return m.Err
}

func (m *MockDataStore) GetMaskedAPIKeys() (map[string]string, error) {
	if m.GetMaskedAPIKeysFunc != nil {
		return m.GetMaskedAPIKeysFunc()
	}
	if m.Err != nil {
		return nil, m.Err
	}
	return map[string]string{}, nil
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

	_, err = mock.GetMaskedAPIKeys()
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
	ds.WritePrompts([]Prompt{{Text: "P1"}, {Text: "P2"}})

	// Update order - should not panic
	ds.UpdatePromptsOrder([]int{1, 0})
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
	ds.WritePrompts([]Prompt{{Text: "P1"}})

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

func TestSQLiteDataStore_GetSetSetting(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}

	// Set a setting
	err = ds.SetSetting("test_key", "test_value")
	if err != nil {
		t.Errorf("SetSetting failed: %v", err)
	}

	// Get it back
	val, err := ds.GetSetting("test_key")
	if err != nil {
		t.Errorf("GetSetting failed: %v", err)
	}
	if val != "test_value" {
		t.Errorf("expected 'test_value', got %q", val)
	}
}

func TestSQLiteDataStore_GetSetAPIKey(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Set encryption key for this test
	t.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	ds := &SQLiteDataStore{}

	// Set API key
	err = ds.SetAPIKey("openai", "sk-test-key")
	if err != nil {
		t.Errorf("SetAPIKey failed: %v", err)
	}

	// Get it back
	key, err := ds.GetAPIKey("openai")
	if err != nil {
		t.Errorf("GetAPIKey failed: %v", err)
	}
	if key != "sk-test-key" {
		t.Errorf("expected 'sk-test-key', got %q", key)
	}
}

func TestSQLiteDataStore_GetMaskedAPIKeys(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	ds := &SQLiteDataStore{}

	masked, err := ds.GetMaskedAPIKeys()
	if err != nil {
		t.Errorf("GetMaskedAPIKeys failed: %v", err)
	}
	if masked == nil {
		t.Error("expected non-nil masked keys map")
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

	// Should not panic
	ds.BroadcastResults()
}
