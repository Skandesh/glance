package glance

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptionService handles encryption and decryption of sensitive data like API keys
type EncryptionService struct {
	key    []byte
	mu     sync.RWMutex
	cached sync.Map // Cache for encrypted values to avoid repeated encryption
}

var (
	globalEncryption     *EncryptionService
	globalEncryptionOnce sync.Once
)

// GetEncryptionService returns the global encryption service (singleton)
func GetEncryptionService() (*EncryptionService, error) {
	var initErr error
	globalEncryptionOnce.Do(func() {
		masterKey := os.Getenv("GLANCE_MASTER_KEY")
		if masterKey == "" {
			// Generate a warning but allow operation
			// In production, GLANCE_MASTER_KEY should always be set
			masterKey = generateDefaultKey()
		}

		// Derive encryption key using PBKDF2
		salt := []byte("glance-business-dashboard-salt-v1")
		key := pbkdf2.Key([]byte(masterKey), salt, 100000, 32, sha256.New)

		globalEncryption = &EncryptionService{
			key: key,
		}
	})

	return globalEncryption, initErr
}

// generateDefaultKey generates a default key for development (NOT FOR PRODUCTION)
func generateDefaultKey() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("glance-dev-key-%s", hostname)
}

// Encrypt encrypts plaintext using AES-256-GCM
func (e *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Check cache
	if cached, ok := e.cached.Load(plaintext); ok {
		return cached.(string), nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// Cache the result
	e.cached.Store(plaintext, encoded)

	return encoded, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (e *EncryptionService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptIfNeeded encrypts a value if it doesn't start with "encrypted:"
func (e *EncryptionService) EncryptIfNeeded(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	// Check if already encrypted
	if len(value) > 10 && value[:10] == "encrypted:" {
		return value, nil
	}

	encrypted, err := e.Encrypt(value)
	if err != nil {
		return "", err
	}

	return "encrypted:" + encrypted, nil
}

// DecryptIfNeeded decrypts a value if it starts with "encrypted:"
func (e *EncryptionService) DecryptIfNeeded(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	// Check if encrypted
	if len(value) > 10 && value[:10] == "encrypted:" {
		return e.Decrypt(value[10:])
	}

	// Return as-is if not encrypted (for backward compatibility)
	return value, nil
}

// SecureString is a type that prevents accidental logging of sensitive data
type SecureString struct {
	value string
}

// NewSecureString creates a new SecureString
func NewSecureString(value string) *SecureString {
	return &SecureString{value: value}
}

// Get returns the actual value
func (s *SecureString) Get() string {
	return s.value
}

// String returns a masked version for logging
func (s *SecureString) String() string {
	if len(s.value) <= 8 {
		return "***"
	}
	return s.value[:4] + "..." + s.value[len(s.value)-4:]
}

// MarshalJSON prevents the value from being serialized
func (s *SecureString) MarshalJSON() ([]byte, error) {
	return []byte(`"***"`), nil
}

// ValidateAPIKey validates that an API key has the correct format
func ValidateAPIKey(key string, expectedPrefix string) error {
	if key == "" {
		return fmt.Errorf("API key is empty")
	}

	if len(key) < 20 {
		return fmt.Errorf("API key is too short (minimum 20 characters)")
	}

	if expectedPrefix != "" {
		if len(key) < len(expectedPrefix) || key[:len(expectedPrefix)] != expectedPrefix {
			return fmt.Errorf("API key must start with '%s'", expectedPrefix)
		}
	}

	return nil
}

// SanitizeAPIKeyForLogs returns a safe version of an API key for logging
func SanitizeAPIKeyForLogs(key string) string {
	if key == "" {
		return "<empty>"
	}

	if len(key) <= 12 {
		return "***"
	}

	return key[:8] + "..." + key[len(key)-4:]
}
