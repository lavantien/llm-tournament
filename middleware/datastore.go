package middleware

// DataStore defines the interface for data persistence operations
type DataStore interface {
	// Suite operations
	GetCurrentSuiteID() (int, error)
	GetCurrentSuiteName() string
	ListSuites() ([]string, error)
	SetCurrentSuite(name string) error
	SuiteExists(name string) bool

	// Prompt operations
	ReadPrompts() []Prompt
	WritePrompts(prompts []Prompt) error
	ReadPromptSuite(suiteName string) ([]Prompt, error)
	WritePromptSuite(suiteName string, prompts []Prompt) error
	ListPromptSuites() ([]string, error)
	UpdatePromptsOrder(order []int)

	// Profile operations
	ReadProfiles() []Profile
	WriteProfiles(profiles []Profile) error

	// Results operations
	ReadResults() map[string]Result
	WriteResults(suiteName string, results map[string]Result) error

	// Settings operations
	GetSetting(key string) (string, error)
	SetSetting(key, value string) error
	GetAPIKey(provider string) (string, error)
	SetAPIKey(provider, key string) error
	GetMaskedAPIKeys() (map[string]string, error)

	// Broadcast
	BroadcastResults()
}

// SQLiteDataStore implements DataStore using SQLite
type SQLiteDataStore struct{}

// GetCurrentSuiteID delegates to the package-level function
func (s *SQLiteDataStore) GetCurrentSuiteID() (int, error) {
	return GetCurrentSuiteID()
}

// GetCurrentSuiteName delegates to the package-level function
func (s *SQLiteDataStore) GetCurrentSuiteName() string {
	return GetCurrentSuiteName()
}

// ListSuites delegates to the package-level function
func (s *SQLiteDataStore) ListSuites() ([]string, error) {
	return ListSuites()
}

// SetCurrentSuite delegates to the package-level function
func (s *SQLiteDataStore) SetCurrentSuite(name string) error {
	return SetCurrentSuite(name)
}

// SuiteExists delegates to the package-level function
func (s *SQLiteDataStore) SuiteExists(name string) bool {
	return SuiteExists(name)
}

// ReadPrompts delegates to the package-level function
func (s *SQLiteDataStore) ReadPrompts() []Prompt {
	return ReadPrompts()
}

// WritePrompts delegates to the package-level function
func (s *SQLiteDataStore) WritePrompts(prompts []Prompt) error {
	return WritePrompts(prompts)
}

// ReadPromptSuite delegates to the package-level function
func (s *SQLiteDataStore) ReadPromptSuite(suiteName string) ([]Prompt, error) {
	return ReadPromptSuite(suiteName)
}

// WritePromptSuite delegates to the package-level function
func (s *SQLiteDataStore) WritePromptSuite(suiteName string, prompts []Prompt) error {
	return WritePromptSuite(suiteName, prompts)
}

// ListPromptSuites delegates to the package-level function
func (s *SQLiteDataStore) ListPromptSuites() ([]string, error) {
	return ListPromptSuites()
}

// UpdatePromptsOrder delegates to the package-level function
func (s *SQLiteDataStore) UpdatePromptsOrder(order []int) {
	UpdatePromptsOrder(order)
}

// ReadProfiles delegates to the package-level function
func (s *SQLiteDataStore) ReadProfiles() []Profile {
	return ReadProfiles()
}

// WriteProfiles delegates to the package-level function
func (s *SQLiteDataStore) WriteProfiles(profiles []Profile) error {
	return WriteProfiles(profiles)
}

// ReadResults delegates to the package-level function
func (s *SQLiteDataStore) ReadResults() map[string]Result {
	return ReadResults()
}

// WriteResults delegates to the package-level function
func (s *SQLiteDataStore) WriteResults(suiteName string, results map[string]Result) error {
	return WriteResults(suiteName, results)
}

// GetSetting delegates to the package-level function
func (s *SQLiteDataStore) GetSetting(key string) (string, error) {
	return GetSetting(key)
}

// SetSetting delegates to the package-level function
func (s *SQLiteDataStore) SetSetting(key, value string) error {
	return SetSetting(key, value)
}

// GetAPIKey delegates to the package-level function
func (s *SQLiteDataStore) GetAPIKey(provider string) (string, error) {
	return GetAPIKey(provider)
}

// SetAPIKey delegates to the package-level function
func (s *SQLiteDataStore) SetAPIKey(provider, key string) error {
	return SetAPIKey(provider, key)
}

// GetMaskedAPIKeys delegates to the package-level function
func (s *SQLiteDataStore) GetMaskedAPIKeys() (map[string]string, error) {
	return GetMaskedAPIKeys()
}

// BroadcastResults delegates to the package-level function
func (s *SQLiteDataStore) BroadcastResults() {
	BroadcastResults()
}

// DefaultDataStore is the default DataStore instance
var DefaultDataStore DataStore = &SQLiteDataStore{}
