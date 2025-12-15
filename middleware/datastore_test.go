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
