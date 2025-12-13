package middleware

import (
	"database/sql"
	"fmt"
	"time"
)

// GetSetting retrieves a setting value (encrypted values are not decrypted)
func GetSetting(key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetSetting updates or inserts a setting
func SetSetting(key, value string) error {
	now := time.Now()
	_, err := db.Exec(`
		INSERT OR REPLACE INTO settings (key, value, created_at, updated_at)
		VALUES (?, ?, COALESCE((SELECT created_at FROM settings WHERE key = ?), ?), ?)
	`, key, value, key, now, now)
	return err
}

// GetAPIKey retrieves and decrypts an API key
func GetAPIKey(provider string) (string, error) {
	encrypted, err := GetSetting(fmt.Sprintf("api_key_%s", provider))
	if err != nil || encrypted == "" {
		return "", err
	}
	return DecryptAPIKey(encrypted)
}

// SetAPIKey encrypts and stores an API key
func SetAPIKey(provider, apiKey string) error {
	encrypted, err := EncryptAPIKey(apiKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt API key: %w", err)
	}
	return SetSetting(fmt.Sprintf("api_key_%s", provider), encrypted)
}

// GetAllSettings retrieves all settings as a map
func GetAllSettings() (map[string]string, error) {
	rows, err := db.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}
	return settings, nil
}

// GetMaskedAPIKeys retrieves all API keys in masked form for display
func GetMaskedAPIKeys() (map[string]string, error) {
	rows, err := db.Query("SELECT key, value FROM settings WHERE key LIKE 'api_key_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apiKeys := make(map[string]string)
	for rows.Next() {
		var key, encryptedValue string
		if err := rows.Scan(&key, &encryptedValue); err != nil {
			return nil, err
		}

		// Decrypt and mask
		decrypted, err := DecryptAPIKey(encryptedValue)
		if err != nil {
			apiKeys[key] = "***ERROR***"
			continue
		}

		apiKeys[key] = MaskAPIKey(decrypted)
	}

	return apiKeys, nil
}
