package testutil

import (
	"database/sql"
	"html/template"
	"net/http"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var (
	sqlOpen           = sql.Open
	enableForeignKeys = func(db *sql.DB) error {
		_, err := db.Exec("PRAGMA foreign_keys = ON")
		return err
	}
)

var (
	createTestSchemaFunc = createTestSchema
	lastInsertID         = func(result sql.Result) (int64, error) { return result.LastInsertId() }
	fatalf               = (*testing.T).Fatalf
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

// SetupTestDB creates an in-memory SQLite database with schema for testing
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sqlOpen("sqlite3", ":memory:")
	if err != nil {
		fatalf(t, "failed to open test database: %v", err)
		return nil
	}

	// Enable foreign keys
	err = enableForeignKeys(db)
	if err != nil {
		_ = db.Close()
		fatalf(t, "failed to enable foreign keys: %v", err)
		return nil
	}

	// Create schema
	if err := createTestSchemaFunc(db); err != nil {
		_ = db.Close()
		fatalf(t, "failed to create test schema: %v", err)
		return nil
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

		CREATE TABLE IF NOT EXISTS model_responses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_id INTEGER NOT NULL,
			prompt_id INTEGER NOT NULL,
			response_text TEXT DEFAULT '',
			response_source TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
			FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
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
		fatalf(t, "failed to create test suite: %v", err)
		return 0
	}
	id, err := lastInsertID(result)
	if err != nil {
		fatalf(t, "failed to get suite id: %v", err)
		return 0
	}
	return int(id)
}

// CreateTestProfile creates a test profile and returns its ID
func CreateTestProfile(t *testing.T, db *sql.DB, suiteID int, name, description string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO profiles (name, description, suite_id) VALUES (?, ?, ?)", name, description, suiteID)
	if err != nil {
		fatalf(t, "failed to create test profile: %v", err)
		return 0
	}
	id, err := lastInsertID(result)
	if err != nil {
		fatalf(t, "failed to get profile id: %v", err)
		return 0
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
		fatalf(t, "failed to create test prompt: %v", err)
		return 0
	}
	id, err := lastInsertID(result)
	if err != nil {
		fatalf(t, "failed to get prompt id: %v", err)
		return 0
	}
	return int(id)
}

// CreateTestModel creates a test model and returns its ID
func CreateTestModel(t *testing.T, db *sql.DB, suiteID int, name string) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)", name, suiteID)
	if err != nil {
		fatalf(t, "failed to create test model: %v", err)
		return 0
	}
	id, err := lastInsertID(result)
	if err != nil {
		fatalf(t, "failed to get model id: %v", err)
		return 0
	}
	return int(id)
}

// CreateTestScore creates a test score
func CreateTestScore(t *testing.T, db *sql.DB, modelID, promptID, score int) {
	t.Helper()
	_, err := db.Exec("INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?)", modelID, promptID, score)
	if err != nil {
		fatalf(t, "failed to create test score: %v", err)
		return
	}
}

// GetDefaultSuiteID returns the ID of the default suite
func GetDefaultSuiteID(t *testing.T, db *sql.DB) int {
	t.Helper()
	var id int
	err := db.QueryRow("SELECT id FROM suites WHERE name = 'default'").Scan(&id)
	if err != nil {
		fatalf(t, "failed to get default suite id: %v", err)
		return 0
	}
	return id
}

// SetCurrentSuite sets the current suite
func SetCurrentSuite(t *testing.T, db *sql.DB, suiteID int) {
	t.Helper()
	_, err := db.Exec("UPDATE suites SET is_current = 0")
	if err != nil {
		fatalf(t, "failed to clear current suite: %v", err)
		return
	}
	_, err = db.Exec("UPDATE suites SET is_current = 1 WHERE id = ?", suiteID)
	if err != nil {
		fatalf(t, "failed to set current suite: %v", err)
		return
	}
}

// CreateTestModelResponse creates a test model response
func CreateTestModelResponse(t *testing.T, db *sql.DB, modelID, promptID int, responseText string) {
	t.Helper()
	_, err := db.Exec("INSERT INTO model_responses (model_id, prompt_id, response_text) VALUES (?, ?, ?)",
		modelID, promptID, responseText)
	if err != nil {
		fatalf(t, "failed to create test model response: %v", err)
		return
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
	_, _ = w.Write([]byte("mock rendered"))
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
	BroadcastResultsFunc    func()

	// Default error to return
	Err error

	// Mock data
	Prompts      []Prompt
	Profiles     []Profile
	Results      map[string]Result
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

// BroadcastResults does nothing in mock
func (m *MockDataStore) BroadcastResults() {
	if m.BroadcastResultsFunc != nil {
		m.BroadcastResultsFunc()
	}
}
