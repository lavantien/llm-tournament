package middleware

import (
	"os"
	"strings"
	"testing"

	"llm-tournament/testutil"
)

func TestGetEncryptionKey(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantErr   bool
		errContains string
	}{
		{
			name:      "valid key",
			envValue:  testutil.ValidEncryptionKey(),
			wantErr:   false,
		},
		{
			name:        "missing env var",
			envValue:   "",
			wantErr:    true,
			errContains: "not set",
		},
		{
			name:        "too short",
			envValue:   "0123456789abcdef",
			wantErr:    true,
			errContains: "64 hex characters",
		},
		{
			name:        "too long",
			envValue:   "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef00",
			wantErr:    true,
			errContains: "64 hex characters",
		},
		{
			name:        "invalid hex",
			envValue:   "ghijklmnopqrstuv0123456789abcdef0123456789abcdef0123456789abcdef",
			wantErr:    true,
			errContains: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env
			original := os.Getenv("ENCRYPTION_KEY")
			defer func() {
				if original == "" {
					os.Unsetenv("ENCRYPTION_KEY")
				} else {
					os.Setenv("ENCRYPTION_KEY", original)
				}
			}()

			if tt.envValue == "" {
				os.Unsetenv("ENCRYPTION_KEY")
			} else {
				os.Setenv("ENCRYPTION_KEY", tt.envValue)
			}

			key, err := getEncryptionKey()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(key) != 32 {
					t.Errorf("expected 32-byte key, got %d bytes", len(key))
				}
			}
		})
	}
}

func TestEncryptAPIKey(t *testing.T) {
	cleanup := testutil.SetupEncryptionKey(t)
	defer cleanup()

	tests := []struct {
		name      string
		plaintext string
		wantEmpty bool
	}{
		{
			name:      "empty string passthrough",
			plaintext: "",
			wantEmpty: true,
		},
		{
			name:      "simple key",
			plaintext: "sk-1234567890",
		},
		{
			name:      "long key",
			plaintext: "sk-proj-1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnop",
		},
		{
			name:      "special characters",
			plaintext: "sk-proj-ABC!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "unicode characters",
			plaintext: "sk-测试密钥-🔑",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptAPIKey(tt.plaintext)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantEmpty {
				if encrypted != "" {
					t.Errorf("expected empty string, got %q", encrypted)
				}
			} else {
				if encrypted == "" {
					t.Error("expected non-empty ciphertext")
				}
				if encrypted == tt.plaintext {
					t.Error("ciphertext should differ from plaintext")
				}
			}
		})
	}
}

func TestEncryptAPIKey_NoKey(t *testing.T) {
	cleanup := testutil.ClearEncryptionKey(t)
	defer cleanup()

	_, err := EncryptAPIKey("test-key")
	if err == nil {
		t.Error("expected error when ENCRYPTION_KEY not set")
	}
}

func TestDecryptAPIKey(t *testing.T) {
	cleanup := testutil.SetupEncryptionKey(t)
	defer cleanup()

	tests := []struct {
		name       string
		ciphertext string
		wantEmpty  bool
		wantErr    bool
	}{
		{
			name:       "empty string passthrough",
			ciphertext: "",
			wantEmpty:  true,
		},
		{
			name:       "invalid base64",
			ciphertext: "not-valid-base64!!!",
			wantErr:    true,
		},
		{
			name:       "too short ciphertext",
			ciphertext: "YWJj", // "abc" in base64
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := DecryptAPIKey(tt.ciphertext)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.wantEmpty && decrypted != "" {
					t.Errorf("expected empty string, got %q", decrypted)
				}
			}
		})
	}
}

func TestDecryptAPIKey_NoKey(t *testing.T) {
	cleanup := testutil.ClearEncryptionKey(t)
	defer cleanup()

	_, err := DecryptAPIKey("somebase64data==")
	if err == nil {
		t.Error("expected error when ENCRYPTION_KEY not set")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	cleanup := testutil.SetupEncryptionKey(t)
	defer cleanup()

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple key", "sk-1234567890"},
		{"empty string", ""},
		{"long key", "sk-proj-1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnop"},
		{"special characters", "sk-proj-ABC!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"unicode", "sk-测试密钥-🔑"},
		{"newlines", "key-with\nnewlines\nand\ttabs"},
		{"spaces", "key with spaces in it"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptAPIKey(tt.plaintext)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			decrypted, err := DecryptAPIKey(encrypted)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("round-trip failed: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	cleanup := testutil.SetupEncryptionKey(t)
	defer cleanup()

	plaintext := "sk-test-key-12345"

	// Encrypt the same plaintext twice
	encrypted1, err := EncryptAPIKey(plaintext)
	if err != nil {
		t.Fatalf("first encrypt error: %v", err)
	}

	encrypted2, err := EncryptAPIKey(plaintext)
	if err != nil {
		t.Fatalf("second encrypt error: %v", err)
	}

	// Due to random nonce, ciphertexts should be different
	if encrypted1 == encrypted2 {
		t.Error("expected different ciphertexts due to random nonce")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := DecryptAPIKey(encrypted1)
	if err != nil {
		t.Fatalf("first decrypt error: %v", err)
	}

	decrypted2, err := DecryptAPIKey(encrypted2)
	if err != nil {
		t.Fatalf("second decrypt error: %v", err)
	}

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("both ciphertexts should decrypt to original plaintext")
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{
			name:   "empty string",
			apiKey: "",
			want:   "",
		},
		{
			name:   "single character",
			apiKey: "a",
			want:   "****",
		},
		{
			name:   "two characters",
			apiKey: "ab",
			want:   "****",
		},
		{
			name:   "three characters",
			apiKey: "abc",
			want:   "****",
		},
		{
			name:   "four characters",
			apiKey: "abcd",
			want:   "****",
		},
		{
			name:   "five characters",
			apiKey: "abcde",
			want:   "sk-...bcde",
		},
		{
			name:   "normal API key",
			apiKey: "sk-1234567890abcdef",
			want:   "sk-...cdef",
		},
		{
			name:   "long API key",
			apiKey: "sk-proj-1234567890abcdefghijklmnopqrstuvwxyz",
			want:   "sk-...wxyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskAPIKey(tt.apiKey)
			if got != tt.want {
				t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.apiKey, got, tt.want)
			}
		})
	}
}
