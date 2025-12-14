package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"llm-tournament/middleware"

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
