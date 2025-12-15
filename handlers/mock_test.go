package handlers

import (
	"html/template"
	"net/http"

	"llm-tournament/middleware"
)

// MockDataStore implements middleware.DataStore for handler testing with error injection
type MockDataStore struct {
	// Function hooks for custom behavior
	WriteResultsFunc     func(suiteName string, results map[string]middleware.Result) error
	WritePromptsFunc     func(prompts []middleware.Prompt) error
	WriteProfilesFunc    func(profiles []middleware.Profile) error
	BroadcastResultsFunc func()
	GetMaskedAPIKeysFunc func() (map[string]string, error)
	SetAPIKeyFunc        func(provider, key string) error

	// Mock data
	Prompts      []middleware.Prompt
	Profiles     []middleware.Profile
	Results      map[string]middleware.Result
	Settings     map[string]string
	CurrentSuite string
}

func (m *MockDataStore) GetCurrentSuiteID() (int, error) { return 1, nil }
func (m *MockDataStore) GetCurrentSuiteName() string {
	if m.CurrentSuite != "" {
		return m.CurrentSuite
	}
	return "default"
}
func (m *MockDataStore) ListSuites() ([]string, error)          { return []string{"default"}, nil }
func (m *MockDataStore) SetCurrentSuite(name string) error      { return nil }
func (m *MockDataStore) SuiteExists(name string) bool           { return true }
func (m *MockDataStore) ReadPrompts() []middleware.Prompt       { return m.Prompts }
func (m *MockDataStore) WritePrompts(prompts []middleware.Prompt) error {
	if m.WritePromptsFunc != nil {
		return m.WritePromptsFunc(prompts)
	}
	m.Prompts = prompts
	return nil
}
func (m *MockDataStore) ReadPromptSuite(suiteName string) ([]middleware.Prompt, error) {
	return m.Prompts, nil
}
func (m *MockDataStore) WritePromptSuite(suiteName string, prompts []middleware.Prompt) error {
	m.Prompts = prompts
	return nil
}
func (m *MockDataStore) ListPromptSuites() ([]string, error) { return []string{"default"}, nil }
func (m *MockDataStore) UpdatePromptsOrder(order []int)      {}
func (m *MockDataStore) ReadProfiles() []middleware.Profile  { return m.Profiles }
func (m *MockDataStore) WriteProfiles(profiles []middleware.Profile) error {
	if m.WriteProfilesFunc != nil {
		return m.WriteProfilesFunc(profiles)
	}
	m.Profiles = profiles
	return nil
}
func (m *MockDataStore) ReadResults() map[string]middleware.Result {
	if m.Results == nil {
		return make(map[string]middleware.Result)
	}
	return m.Results
}
func (m *MockDataStore) WriteResults(suiteName string, results map[string]middleware.Result) error {
	if m.WriteResultsFunc != nil {
		return m.WriteResultsFunc(suiteName, results)
	}
	m.Results = results
	return nil
}
func (m *MockDataStore) GetSetting(key string) (string, error) {
	if m.Settings != nil {
		return m.Settings[key], nil
	}
	return "", nil
}
func (m *MockDataStore) SetSetting(key, value string) error {
	if m.Settings == nil {
		m.Settings = make(map[string]string)
	}
	m.Settings[key] = value
	return nil
}
func (m *MockDataStore) GetAPIKey(provider string) (string, error) { return "", nil }
func (m *MockDataStore) SetAPIKey(provider, key string) error {
	if m.SetAPIKeyFunc != nil {
		return m.SetAPIKeyFunc(provider, key)
	}
	return nil
}
func (m *MockDataStore) GetMaskedAPIKeys() (map[string]string, error) {
	if m.GetMaskedAPIKeysFunc != nil {
		return m.GetMaskedAPIKeysFunc()
	}
	return map[string]string{}, nil
}
func (m *MockDataStore) BroadcastResults() {
	if m.BroadcastResultsFunc != nil {
		m.BroadcastResultsFunc()
	}
}

// MockRenderer implements middleware.TemplateRenderer for testing
type MockRenderer struct {
	RenderError error
}

func (m *MockRenderer) Render(w http.ResponseWriter, name string, funcMap template.FuncMap, data interface{}, files ...string) error {
	if m.RenderError != nil {
		return m.RenderError
	}
	w.Write([]byte("mock rendered"))
	return nil
}

func (m *MockRenderer) RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error {
	return m.Render(w, tmpl, nil, data, "templates/"+tmpl)
}

// FailingResponseWriter is a ResponseWriter that can simulate write failures
type FailingResponseWriter struct {
	http.ResponseWriter
	WriteError    error
	HeaderWritten bool
}

func (f *FailingResponseWriter) Write(p []byte) (int, error) {
	if f.WriteError != nil {
		return 0, f.WriteError
	}
	return f.ResponseWriter.Write(p)
}

func (f *FailingResponseWriter) Header() http.Header {
	return f.ResponseWriter.Header()
}

func (f *FailingResponseWriter) WriteHeader(statusCode int) {
	f.HeaderWritten = true
	f.ResponseWriter.WriteHeader(statusCode)
}

// MockDataStoreWithError creates a MockDataStore that returns specified errors
type MockDataStoreWithError struct {
	MockDataStore
	GetCurrentSuiteIDErr   error
	ReadPromptSuiteErr     error
	WritePromptSuiteErr    error
	ReadProfileSuiteErr    error
	WriteResultsErr        error
	ReadResultsErr         error
}

func (m *MockDataStoreWithError) GetCurrentSuiteID() (int, error) {
	if m.GetCurrentSuiteIDErr != nil {
		return 0, m.GetCurrentSuiteIDErr
	}
	return m.MockDataStore.GetCurrentSuiteID()
}

func (m *MockDataStoreWithError) ReadPromptSuite(suiteName string) ([]middleware.Prompt, error) {
	if m.ReadPromptSuiteErr != nil {
		return nil, m.ReadPromptSuiteErr
	}
	return m.MockDataStore.ReadPromptSuite(suiteName)
}

func (m *MockDataStoreWithError) WritePromptSuite(suiteName string, prompts []middleware.Prompt) error {
	if m.WritePromptSuiteErr != nil {
		return m.WritePromptSuiteErr
	}
	return m.MockDataStore.WritePromptSuite(suiteName, prompts)
}

func (m *MockDataStoreWithError) WriteResults(suiteName string, results map[string]middleware.Result) error {
	if m.WriteResultsErr != nil {
		return m.WriteResultsErr
	}
	return m.MockDataStore.WriteResults(suiteName, results)
}
