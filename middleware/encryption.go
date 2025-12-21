package middleware

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

var (
	aesNewCipher           = aes.NewCipher
	cipherNewGCM           = cipher.NewGCM
	randReader   io.Reader = rand.Reader
)

// getEncryptionKey retrieves the encryption key from environment
func getEncryptionKey() ([]byte, error) {
	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY environment variable not set")
	}

	// Convert hex string to bytes (expecting 32-byte hex = 64 characters)
	if len(keyHex) != 64 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be 64 hex characters (32 bytes)")
	}

	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		_, err := fmt.Sscanf(keyHex[i*2:i*2+2], "%02x", &key[i])
		if err != nil {
			return nil, fmt.Errorf("invalid ENCRYPTION_KEY format: %w", err)
		}
	}

	return key, nil
}

// EncryptAPIKey encrypts an API key using AES-256-GCM
func EncryptAPIKey(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aesNewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipherNewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(randReader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAPIKey decrypts an API key using AES-256-GCM
func DecryptAPIKey(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aesNewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipherNewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// MaskAPIKey masks an API key for display (show only last 4 characters)
func MaskAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}

	if len(apiKey) <= 4 {
		return "****"
	}

	return "sk-..." + apiKey[len(apiKey)-4:]
}
