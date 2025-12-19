package testutil

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Prompt is a local type matching middleware.Prompt for testing
type Prompt struct {
	Text         string
	Solution     string
	Profile      string
	DisplayOrder int
	Type         string
}

// Profile is a local type matching middleware.Profile for testing
type Profile struct {
	Name        string
	Description string
}

// Result is a local type matching middleware.Result for testing
type Result struct {
	Scores []int
}

// ValidEncryptionKey returns a valid 64-char hex key for testing
func ValidEncryptionKey() string {
	return "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
}

// SetupEncryptionKey sets up a valid ENCRYPTION_KEY env var and returns cleanup function
func SetupEncryptionKey(t *testing.T) func() {
	t.Helper()
	original := os.Getenv("ENCRYPTION_KEY")
	os.Setenv("ENCRYPTION_KEY", ValidEncryptionKey())

	return func() {
		if original == "" {
			os.Unsetenv("ENCRYPTION_KEY")
		} else {
			os.Setenv("ENCRYPTION_KEY", original)
		}
	}
}

// ClearEncryptionKey removes the ENCRYPTION_KEY env var and returns cleanup function
func ClearEncryptionKey(t *testing.T) func() {
	t.Helper()
	original := os.Getenv("ENCRYPTION_KEY")
	os.Unsetenv("ENCRYPTION_KEY")

	return func() {
		if original != "" {
			os.Setenv("ENCRYPTION_KEY", original)
		}
	}
}

// SetupTestDB creates an in-memory SQLite database with schema for testing
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Create schema
	if err := createTestSchema(db); err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}

	return db
}

// createTestSchema creates the database schema for testing
func createTestSchema(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS suites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			is_current INTEGER DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			suite_id INTEGER NOT NULL,
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
			UNIQUE(name, suite_id)
		);

		CREATE TABLE IF NOT EXISTS prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			solution TEXT DEFAULT '',
			profile_id INTEGER,
			suite_id INTEGER NOT NULL,
			display_order INTEGER DEFAULT 0,
			type TEXT DEFAULT 'objective',
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS models (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			suite_id INTEGER NOT NULL,
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
			UNIQUE(name, suite_id)
		);

		CREATE TABLE IF NOT EXISTS scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_id INTEGER NOT NULL,
			prompt_id INTEGER NOT NULL,
			score INTEGER DEFAULT 0,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
			FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE,
			UNIQUE(model_id, prompt_id)
		);

		CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			value TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS evaluation_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			suite_id INTEGER NOT NULL,
			job_type TEXT NOT NULL,
			target_id INTEGER,
			status TEXT DEFAULT 'pending',
			progress_current INTEGER DEFAULT 0,
			progress_total INTEGER DEFAULT 0,
			estimated_cost_usd REAL DEFAULT 0,
			actual_cost_usd REAL DEFAULT 0,
			error_message TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			completed_at DATETIME,
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS model_responses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_id INTEGER NOT NULL,
			prompt_id INTEGER NOT NULL,
			response_text TEXT DEFAULT '',
			response_source TEXT DEFAULT '',
			api_config TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
			FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS evaluation_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			job_id INTEGER NOT NULL,
			model_id INTEGER NOT NULL,
			prompt_id INTEGER NOT NULL,
			judge_name TEXT NOT NULL,
			judge_score INTEGER DEFAULT 0,
			judge_confidence REAL DEFAULT 0,
			judge_reasoning TEXT DEFAULT '',
			cost_usd REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (job_id) REFERENCES evaluation_jobs(id) ON DELETE CASCADE,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
			FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS cost_tracking (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			suite_id INTEGER NOT NULL,
			date DATE NOT NULL,
			total_cost_usd REAL DEFAULT 0,
			evaluation_count INTEGER DEFAULT 0,
			FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE,
			UNIQUE(suite_id, date)
		);

		-- Insert default suite
		INSERT INTO suites (name, is_current) VALUES ('default', 1);
	`

	_, err := db.Exec(schema)
	return err
}

// CreateTestSuite creates a test suite and returns its ID
func CreateTestSuite(t *testing.T, db *sql.DB, name string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO suites (name, is_current) VALUES (?, 0)", name)
	if err != nil {
		t.Fatalf("failed to create test suite: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get suite id: %v", err)
	}
	return int(id)
}

// CreateTestProfile creates a test profile and returns its ID
func CreateTestProfile(t *testing.T, db *sql.DB, suiteID int, name, description string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO profiles (name, description, suite_id) VALUES (?, ?, ?)", name, description, suiteID)
	if err != nil {
		t.Fatalf("failed to create test profile: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get profile id: %v", err)
	}
	return int(id)
}

// CreateTestPrompt creates a test prompt and returns its ID
func CreateTestPrompt(t *testing.T, db *sql.DB, suiteID int, text, solution string, profileID *int, displayOrder int, promptType string) int {
	t.Helper()
	var result sql.Result
	var err error
	if profileID != nil {
		result, err = db.Exec("INSERT INTO prompts (text, solution, profile_id, suite_id, display_order, type) VALUES (?, ?, ?, ?, ?, ?)",
			text, solution, *profileID, suiteID, displayOrder, promptType)
	} else {
		result, err = db.Exec("INSERT INTO prompts (text, solution, suite_id, display_order, type) VALUES (?, ?, ?, ?, ?)",
			text, solution, suiteID, displayOrder, promptType)
	}
	if err != nil {
		t.Fatalf("failed to create test prompt: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get prompt id: %v", err)
	}
	return int(id)
}

// CreateTestModel creates a test model and returns its ID
func CreateTestModel(t *testing.T, db *sql.DB, suiteID int, name string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", name, suiteID)
	if err != nil {
		t.Fatalf("failed to create test model: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get model id: %v", err)
	}
	return int(id)
}

// CreateTestScore creates a test score
func CreateTestScore(t *testing.T, db *sql.DB, modelID, promptID, score int) {
	t.Helper()
	_, err := db.Exec("INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?)", modelID, promptID, score)
	if err != nil {
		t.Fatalf("failed to create test score: %v", err)
	}
}

// GetDefaultSuiteID returns the ID of the default suite
func GetDefaultSuiteID(t *testing.T, db *sql.DB) int {
	t.Helper()
	var id int
	err := db.QueryRow("SELECT id FROM suites WHERE name = 'default'").Scan(&id)
	if err != nil {
		t.Fatalf("failed to get default suite id: %v", err)
	}
	return id
}

// SetCurrentSuite sets the current suite
func SetCurrentSuite(t *testing.T, db *sql.DB, suiteID int) {
	t.Helper()
	_, err := db.Exec("UPDATE suites SET is_current = 0")
	if err != nil {
		t.Fatalf("failed to clear current suite: %v", err)
	}
	_, err = db.Exec("UPDATE suites SET is_current = 1 WHERE id = ?", suiteID)
	if err != nil {
		t.Fatalf("failed to set current suite: %v", err)
	}
}

// CreateTestSetting creates a test setting
func CreateTestSetting(t *testing.T, db *sql.DB, key, value string) {
	t.Helper()
	_, err := db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	if err != nil {
		t.Fatalf("failed to create test setting: %v", err)
	}
}

// CreateTestEvaluationJob creates a test evaluation job and returns its ID
func CreateTestEvaluationJob(t *testing.T, db *sql.DB, suiteID int, jobType, status string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO evaluation_jobs (suite_id, job_type, status) VALUES (?, ?, ?)", suiteID, jobType, status)
	if err != nil {
		t.Fatalf("failed to create test evaluation job: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get job id: %v", err)
	}
	return int(id)
}

// CreateTestModelResponse creates a test model response
func CreateTestModelResponse(t *testing.T, db *sql.DB, modelID, promptID int, responseText string) {
	t.Helper()
	_, err := db.Exec("INSERT INTO model_responses (model_id, prompt_id, response_text) VALUES (?, ?, ?)",
		modelID, promptID, responseText)
	if err != nil {
		t.Fatalf("failed to create test model response: %v", err)
	}
}

// MockRenderer implements TemplateRenderer for testing with error injection
type MockRenderer struct {
	RenderError error
	RenderCalls []MockRenderCall
}

// MockRenderCall records a call to Render
type MockRenderCall struct {
	Name  string
	Data  interface{}
	Files []string
}

// Render records the call and returns any configured error
func (m *MockRenderer) Render(w http.ResponseWriter, name string, funcMap template.FuncMap, data interface{}, files ...string) error {
	m.RenderCalls = append(m.RenderCalls, MockRenderCall{
		Name:  name,
		Data:  data,
		Files: files,
	})
	if m.RenderError != nil {
		return m.RenderError
	}
	// Write minimal content to satisfy tests expecting output
	w.Write([]byte("mock rendered"))
	return nil
}

// RenderTemplateSimple records the call and returns any configured error
func (m *MockRenderer) RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error {
	return m.Render(w, tmpl, nil, data, "templates/"+tmpl)
}

// MockDataStore implements DataStore interface for testing with error injection
// Note: This uses local types (Prompt, Profile, Result) that mirror middleware types
// The handlers tests must use type assertions or conversion when using this mock
type MockDataStore struct {
	// Function hooks for custom behavior
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

	// Default error to return
	Err error

	// Mock data
	Prompts      []Prompt
	Profiles     []Profile
	Results      map[string]Result
	Settings     map[string]string
	CurrentSuite string
}

// GetCurrentSuiteID returns mock suite ID or error
func (m *MockDataStore) GetCurrentSuiteID() (int, error) {
	if m.GetCurrentSuiteIDFunc != nil {
		return m.GetCurrentSuiteIDFunc()
	}
	if m.Err != nil {
		return 0, m.Err
	}
	return 1, nil
}

// GetCurrentSuiteName returns mock suite name
func (m *MockDataStore) GetCurrentSuiteName() string {
	if m.GetCurrentSuiteNameFunc != nil {
		return m.GetCurrentSuiteNameFunc()
	}
	if m.CurrentSuite != "" {
		return m.CurrentSuite
	}
	return "default"
}

// ListSuites returns mock suites or error
func (m *MockDataStore) ListSuites() ([]string, error) {
	if m.ListSuitesFunc != nil {
		return m.ListSuitesFunc()
	}
	if m.Err != nil {
		return nil, m.Err
	}
	return []string{"default"}, nil
}

// SetCurrentSuite returns mock error
func (m *MockDataStore) SetCurrentSuite(name string) error {
	if m.SetCurrentSuiteFunc != nil {
		return m.SetCurrentSuiteFunc(name)
	}
	if m.Err != nil {
		return m.Err
	}
	m.CurrentSuite = name
	return nil
}

// SuiteExists returns mock result
func (m *MockDataStore) SuiteExists(name string) bool {
	if m.SuiteExistsFunc != nil {
		return m.SuiteExistsFunc(name)
	}
	return true
}

// ReadPrompts returns mock prompts
func (m *MockDataStore) ReadPrompts() []Prompt {
	if m.ReadPromptsFunc != nil {
		return m.ReadPromptsFunc()
	}
	return m.Prompts
}

// WritePrompts stores prompts or returns error
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

// ReadPromptSuite returns mock prompts for a suite
func (m *MockDataStore) ReadPromptSuite(suiteName string) ([]Prompt, error) {
	if m.ReadPromptSuiteFunc != nil {
		return m.ReadPromptSuiteFunc(suiteName)
	}
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Prompts, nil
}

// WritePromptSuite stores prompts for a suite
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

// ListPromptSuites returns mock suite list
func (m *MockDataStore) ListPromptSuites() ([]string, error) {
	if m.ListPromptSuitesFunc != nil {
		return m.ListPromptSuitesFunc()
	}
	if m.Err != nil {
		return nil, m.Err
	}
	return []string{"default"}, nil
}

// UpdatePromptsOrder updates prompts order
func (m *MockDataStore) UpdatePromptsOrder(order []int) {
	if m.UpdatePromptsOrderFunc != nil {
		m.UpdatePromptsOrderFunc(order)
	}
}

// ReadProfiles returns mock profiles
func (m *MockDataStore) ReadProfiles() []Profile {
	if m.ReadProfilesFunc != nil {
		return m.ReadProfilesFunc()
	}
	return m.Profiles
}

// WriteProfiles stores profiles or returns error
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

// ReadResults returns mock results
func (m *MockDataStore) ReadResults() map[string]Result {
	if m.ReadResultsFunc != nil {
		return m.ReadResultsFunc()
	}
	if m.Results == nil {
		return make(map[string]Result)
	}
	return m.Results
}

// WriteResults stores results or returns error
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

// GetSetting returns mock setting or error
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

// SetSetting stores setting or returns error
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

// GetAPIKey returns mock API key or error
func (m *MockDataStore) GetAPIKey(provider string) (string, error) {
	if m.GetAPIKeyFunc != nil {
		return m.GetAPIKeyFunc(provider)
	}
	if m.Err != nil {
		return "", m.Err
	}
	return "", nil
}

// SetAPIKey stores API key or returns error
func (m *MockDataStore) SetAPIKey(provider, key string) error {
	if m.SetAPIKeyFunc != nil {
		return m.SetAPIKeyFunc(provider, key)
	}
	return m.Err
}

// GetMaskedAPIKeys returns mock masked API keys
func (m *MockDataStore) GetMaskedAPIKeys() (map[string]string, error) {
	if m.GetMaskedAPIKeysFunc != nil {
		return m.GetMaskedAPIKeysFunc()
	}
	if m.Err != nil {
		return nil, m.Err
	}
	return map[string]string{}, nil
}

// BroadcastResults does nothing in mock
func (m *MockDataStore) BroadcastResults() {
	if m.BroadcastResultsFunc != nil {
		m.BroadcastResultsFunc()
	}
}
