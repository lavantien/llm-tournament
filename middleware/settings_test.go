package middleware

import (
	"llm-tournament/testutil"
	"testing"
)

func TestGetSetting(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Get existing default setting
	value, err := GetSetting("cost_alert_threshold_usd")
	if err != nil {
		t.Fatalf("GetSetting failed: %v", err)
	}
	if value != "100.0" {
		t.Errorf("expected '100.0', got %q", value)
	}
}

func TestGetSetting_NonExistent(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	value, err := GetSetting("non_existent_key")
	if err != nil {
		t.Fatalf("GetSetting failed: %v", err)
	}
	if value != "" {
		t.Errorf("expected empty string for non-existent key, got %q", value)
	}
}

func TestSetSetting(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Set a new setting
	err = SetSetting("test_key", "test_value")
	if err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	// Verify it was set
	value, err := GetSetting("test_key")
	if err != nil {
		t.Fatalf("GetSetting failed: %v", err)
	}
	if value != "test_value" {
		t.Errorf("expected 'test_value', got %q", value)
	}
}

func TestSetSetting_Update(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Set initial value
	err = SetSetting("update_key", "initial")
	if err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	// Update the value
	err = SetSetting("update_key", "updated")
	if err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	// Verify it was updated
	value, err := GetSetting("update_key")
	if err != nil {
		t.Fatalf("GetSetting failed: %v", err)
	}
	if value != "updated" {
		t.Errorf("expected 'updated', got %q", value)
	}
}

func TestGetAPIKey(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Setup encryption key
	encCleanup := testutil.SetupEncryptionKey(t)
	defer encCleanup()

	// Set an API key
	err = SetAPIKey("anthropic", "sk-test-key-12345")
	if err != nil {
		t.Fatalf("SetAPIKey failed: %v", err)
	}

	// Get it back
	key, err := GetAPIKey("anthropic")
	if err != nil {
		t.Fatalf("GetAPIKey failed: %v", err)
	}
	if key != "sk-test-key-12345" {
		t.Errorf("expected 'sk-test-key-12345', got %q", key)
	}
}

func TestGetAPIKey_Empty(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Default API keys are empty strings
	key, err := GetAPIKey("anthropic")
	if err != nil {
		t.Fatalf("GetAPIKey failed: %v", err)
	}
	if key != "" {
		t.Errorf("expected empty string for unset API key, got %q", key)
	}
}

func TestSetAPIKey(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	encCleanup := testutil.SetupEncryptionKey(t)
	defer encCleanup()

	tests := []struct {
		provider string
		apiKey   string
	}{
		{"anthropic", "sk-ant-12345"},
		{"openai", "sk-openai-67890"},
		{"google", "AIza-google-key"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			err := SetAPIKey(tt.provider, tt.apiKey)
			if err != nil {
				t.Fatalf("SetAPIKey failed: %v", err)
			}

			// Verify it's encrypted in the database
			raw, err := GetSetting("api_key_" + tt.provider)
			if err != nil {
				t.Fatalf("GetSetting failed: %v", err)
			}
			if raw == tt.apiKey {
				t.Error("API key should be encrypted, not stored in plain text")
			}

			// Verify we can decrypt it
			decrypted, err := GetAPIKey(tt.provider)
			if err != nil {
				t.Fatalf("GetAPIKey failed: %v", err)
			}
			if decrypted != tt.apiKey {
				t.Errorf("decrypted key mismatch: got %q, want %q", decrypted, tt.apiKey)
			}
		})
	}
}

func TestSetAPIKey_NoEncryptionKey(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Clear encryption key
	clearCleanup := testutil.ClearEncryptionKey(t)
	defer clearCleanup()

	err = SetAPIKey("anthropic", "sk-test")
	if err == nil {
		t.Error("expected error when encryption key not set")
	}
}

func TestGetAllSettings(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	settings, err := GetAllSettings()
	if err != nil {
		t.Fatalf("GetAllSettings failed: %v", err)
	}

	// Check default settings exist
	expectedKeys := []string{
		"api_key_anthropic",
		"api_key_openai",
		"api_key_google",
		"cost_alert_threshold_usd",
		"auto_evaluate_new_models",
		"python_service_url",
	}

	for _, key := range expectedKeys {
		if _, exists := settings[key]; !exists {
			t.Errorf("expected setting %q to exist", key)
		}
	}
}

func TestGetMaskedAPIKeys(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	encCleanup := testutil.SetupEncryptionKey(t)
	defer encCleanup()

	// Set an API key
	err = SetAPIKey("anthropic", "sk-test-key-12345678")
	if err != nil {
		t.Fatalf("SetAPIKey failed: %v", err)
	}

	// Get masked keys
	masked, err := GetMaskedAPIKeys()
	if err != nil {
		t.Fatalf("GetMaskedAPIKeys failed: %v", err)
	}

	// Verify the key is masked
	if maskedKey, ok := masked["api_key_anthropic"]; ok {
		if maskedKey == "sk-test-key-12345678" {
			t.Error("API key should be masked, not in plain text")
		}
		if maskedKey != "sk-...5678" {
			t.Errorf("expected masked key 'sk-...5678', got %q", maskedKey)
		}
	} else {
		t.Error("expected api_key_anthropic in masked keys")
	}
}

func TestGetMaskedAPIKeys_Empty(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Default API keys are empty, which mask to empty string
	masked, err := GetMaskedAPIKeys()
	if err != nil {
		t.Fatalf("GetMaskedAPIKeys failed: %v", err)
	}

	// Empty keys should result in empty masked values
	if maskedKey, ok := masked["api_key_anthropic"]; ok {
		if maskedKey != "" {
			t.Errorf("expected empty string for unset API key, got %q", maskedKey)
		}
	}
}
