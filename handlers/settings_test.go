package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"llm-tournament/middleware"
	"llm-tournament/testutil"

	_ "github.com/mattn/go-sqlite3"
)

// setupSettingsTestDB creates a test database for settings handler tests
func setupSettingsTestDB(t *testing.T) func() {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	err := middleware.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	// Set encryption key for API key tests
	os.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	return func() {
		middleware.CloseDB()
		os.Unsetenv("ENCRYPTION_KEY")
	}
}

func TestUpdateSettingsHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/settings/update", nil)
	rr := httptest.NewRecorder()
	UpdateSettingsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestUpdateSettingsHandler_POST_Success(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("cost_alert_threshold_usd", "50.00")
	form.Add("auto_evaluate_new_models", "on")
	form.Add("python_service_url", "http://localhost:8002")

	req := httptest.NewRequest("POST", "/settings/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	UpdateSettingsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify settings were saved
	threshold, _ := middleware.GetSetting("cost_alert_threshold_usd")
	if threshold != "50.00" {
		t.Errorf("expected threshold '50.00', got %q", threshold)
	}

	autoEval, _ := middleware.GetSetting("auto_evaluate_new_models")
	if autoEval != "true" {
		t.Errorf("expected auto_evaluate 'true', got %q", autoEval)
	}

	pythonURL, _ := middleware.GetSetting("python_service_url")
	if pythonURL != "http://localhost:8002" {
		t.Errorf("expected python_service_url 'http://localhost:8002', got %q", pythonURL)
	}
}

func TestUpdateSettingsHandler_POST_AutoEvalOff(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("cost_alert_threshold_usd", "100.00")
	// Note: not setting auto_evaluate_new_models means it should be "false"

	req := httptest.NewRequest("POST", "/settings/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	UpdateSettingsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	autoEval, _ := middleware.GetSetting("auto_evaluate_new_models")
	if autoEval != "false" {
		t.Errorf("expected auto_evaluate 'false', got %q", autoEval)
	}
}

func TestUpdateSettingsHandler_POST_WithAPIKeys(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("api_key_anthropic", "sk-ant-test-key-123")
	form.Add("api_key_openai", "sk-openai-test-key-456")
	form.Add("api_key_google", "AIza-test-key-789")

	req := httptest.NewRequest("POST", "/settings/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	UpdateSettingsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify API keys were saved (read back decrypted)
	anthropicKey, err := middleware.GetAPIKey("anthropic")
	if err != nil {
		t.Errorf("failed to get anthropic key: %v", err)
	}
	if anthropicKey != "sk-ant-test-key-123" {
		t.Errorf("expected anthropic key 'sk-ant-test-key-123', got %q", anthropicKey)
	}
}

func TestUpdateSettingsHandler_POST_SkipsPlaceholder(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	// First set a key
	if err := middleware.SetAPIKey("anthropic", "original-key"); err != nil {
		t.Fatalf("failed to set initial key: %v", err)
	}

	// Now update with placeholder (should be skipped)
	form := url.Values{}
	form.Add("api_key_anthropic", "********")

	req := httptest.NewRequest("POST", "/settings/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	UpdateSettingsHandler(rr, req)

	// Verify original key is preserved
	key, _ := middleware.GetAPIKey("anthropic")
	if key != "original-key" {
		t.Errorf("expected original key 'original-key', got %q", key)
	}
}

func TestTestAPIKeyHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/settings/test_key", nil)
	rr := httptest.NewRecorder()
	TestAPIKeyHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestTestAPIKeyHandler_MissingProvider(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/settings/test_key", nil)
	rr := httptest.NewRecorder()
	TestAPIKeyHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestTestAPIKeyHandler_KeyNotConfigured(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("provider", "anthropic")

	req := httptest.NewRequest("POST", "/settings/test_key", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	TestAPIKeyHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check that response indicates key not configured
	body := rr.Body.String()
	if !strings.Contains(body, "not configured") {
		t.Errorf("expected 'not configured' in response, got %q", body)
	}
}

func TestTestAPIKeyHandler_KeyConfigured(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	// Set a key first
	if err := middleware.SetAPIKey("anthropic", "test-key"); err != nil {
		t.Fatalf("failed to set key: %v", err)
	}

	form := url.Values{}
	form.Add("provider", "anthropic")

	req := httptest.NewRequest("POST", "/settings/test_key", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	TestAPIKeyHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check that response indicates success
	body := rr.Body.String()
	if !strings.Contains(body, "success\":true") {
		t.Errorf("expected success:true in response, got %q", body)
	}
}

// changeToProjectRootSettings changes to the project root directory for tests that need templates
func changeToProjectRootSettings(t *testing.T) func() {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}
	return func() {
		os.Chdir(originalDir)
	}
}

func TestSettingsHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootSettings(t)
	defer restoreDir()

	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/settings", nil)
	rr := httptest.NewRecorder()
	SettingsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Settings") {
		t.Error("expected 'Settings' in response body")
	}
}

func TestSettingsHandler_GET_WithExistingSettings(t *testing.T) {
	restoreDir := changeToProjectRootSettings(t)
	defer restoreDir()

	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	// Set some existing settings
	middleware.SetSetting("cost_alert_threshold_usd", "75.50")
	middleware.SetSetting("auto_evaluate_new_models", "true")
	middleware.SetSetting("python_service_url", "http://custom:9000")
	middleware.SetAPIKey("anthropic", "test-key-123")

	req := httptest.NewRequest("GET", "/settings", nil)
	rr := httptest.NewRecorder()
	SettingsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Settings") {
		t.Error("expected 'Settings' in response body")
	}
}

func TestSettingsHandler_GET_WithZeroThreshold(t *testing.T) {
	restoreDir := changeToProjectRootSettings(t)
	defer restoreDir()

	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	// Set threshold to zero (should default to 100.0)
	middleware.SetSetting("cost_alert_threshold_usd", "0")

	req := httptest.NewRequest("GET", "/settings", nil)
	rr := httptest.NewRecorder()
	SettingsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestUpdateSettingsHandler_POST_EmptyPythonURL(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	// Set initial python URL
	middleware.SetSetting("python_service_url", "http://initial:8001")

	// Update with empty Python URL (should not change existing)
	form := url.Values{}
	form.Add("cost_alert_threshold_usd", "50.00")
	form.Add("python_service_url", "")

	req := httptest.NewRequest("POST", "/settings/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	UpdateSettingsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Python URL should NOT be updated when empty
	pythonURL, _ := middleware.GetSetting("python_service_url")
	if pythonURL != "http://initial:8001" {
		t.Errorf("python URL should remain unchanged, got %q", pythonURL)
	}
}

func TestUpdateSettingsHandler_POST_EmptyThreshold(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	// Set initial threshold
	middleware.SetSetting("cost_alert_threshold_usd", "100")

	// Update with empty threshold (should not change)
	form := url.Values{}
	form.Add("cost_alert_threshold_usd", "")

	req := httptest.NewRequest("POST", "/settings/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	UpdateSettingsHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestSettingsHandler_GET_RenderError(t *testing.T) {
	cleanup := setupSettingsTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	req := httptest.NewRequest("GET", "/settings", nil)
	rr := httptest.NewRecorder()
	SettingsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestSettingsHandler_GetMaskedAPIKeysError(t *testing.T) {
	mockDS := &MockDataStore{
		GetMaskedAPIKeysFunc: func() (map[string]string, error) {
			return nil, errors.New("mock get masked keys error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	req := httptest.NewRequest("GET", "/settings", nil)
	rr := httptest.NewRecorder()
	handler.Settings(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on GetMaskedAPIKeys error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestUpdateSettingsHandler_SetAPIKeyError(t *testing.T) {
	mockDS := &MockDataStore{
		SetAPIKeyFunc: func(provider, key string) error {
			return errors.New("mock set API key error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	form := url.Values{}
	form.Add("api_key_anthropic", "sk-test-key-123")

	req := httptest.NewRequest("POST", "/settings/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.UpdateSettings(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on SetAPIKey error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

type setSettingErrorDataStore struct {
	MockDataStore
	errorsByKey map[string]error
}

func (ds *setSettingErrorDataStore) SetSetting(key, value string) error {
	if err := ds.errorsByKey[key]; err != nil {
		return err
	}
	return ds.MockDataStore.SetSetting(key, value)
}

func TestUpdateSettings_SetSettingErrorsDoNotFailRequest(t *testing.T) {
	ds := &setSettingErrorDataStore{
		errorsByKey: map[string]error{
			"cost_alert_threshold_usd":   errors.New("threshold error"),
			"auto_evaluate_new_models":  errors.New("auto eval error"),
			"python_service_url":        errors.New("python url error"),
		},
	}
	handler := NewHandlerWithDeps(ds, &MockRenderer{})

	form := url.Values{}
	form.Add("cost_alert_threshold_usd", "123.45")
	form.Add("auto_evaluate_new_models", "on")
	form.Add("python_service_url", "http://example:8001")

	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.UpdateSettings(rr, req)

	// These SetSetting errors are logged but the handler still redirects back to Settings.
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestUpdateSettings_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/update_settings", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler.UpdateSettings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Failed to parse form") {
		t.Fatalf("expected parse form error message, got %q", rr.Body.String())
	}
}
