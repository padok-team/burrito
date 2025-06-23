package storage

import (
	"os"
	"testing"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/stretchr/testify/assert"
)

func TestNewEncryptionManager_WithEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name            string
		envKey          string
		configEnabled   bool
		configKey       string
		expectEncryptor bool
		expectError     bool
	}{
		{
			name:            "encryption enabled with env key set",
			envKey:          "test-env-key",
			configEnabled:   true,
			configKey:       "config-key-should-be-ignored",
			expectEncryptor: true,
			expectError:     false,
		},
		{
			name:            "encryption enabled but no env key",
			envKey:          "",
			configEnabled:   true,
			configKey:       "config-key",
			expectEncryptor: false,
			expectError:     true, // This should return an error
		},
		{
			name:            "encryption disabled with env key set",
			envKey:          "test-env-key",
			configEnabled:   false,
			configKey:       "",
			expectEncryptor: false,
			expectError:     false,
		},
		{
			name:            "encryption disabled and no env key",
			envKey:          "",
			configEnabled:   false,
			configKey:       "",
			expectEncryptor: false,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variable
			if tt.envKey != "" {
				err := os.Setenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY", tt.envKey)
				assert.NoError(t, err)
				defer os.Unsetenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")
			} else {
				os.Unsetenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")
			}

			// Create config
			config := config.EncryptionConfig{
				Enabled: tt.configEnabled,
			}

			// Create encryption manager
			em, err := NewEncryptionManager(config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, em)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, em)

			// Check if encryptor is set as expected
			if tt.expectEncryptor {
				assert.NotNil(t, em.defaultEncryptor, "expected default encryptor to be set")
			} else {
				assert.Nil(t, em.defaultEncryptor, "expected default encryptor to be nil")
			}
		})
	}
}

func TestEncryptionManager_EncryptDecrypt(t *testing.T) {
	// Set environment variable
	envKey := "test-encryption-key-123"
	err := os.Setenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY", envKey)
	assert.NoError(t, err)
	defer os.Unsetenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")

	// Create config with encryption enabled
	config := config.EncryptionConfig{
		Enabled: true,
	}

	// Create encryption manager
	em, err := NewEncryptionManager(config)
	assert.NoError(t, err)

	// Test data
	namespace := "test-namespace"
	plaintext := []byte("sensitive terraform state data")

	// Encrypt
	ciphertext, err := em.Encrypt(namespace, plaintext)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext, "ciphertext should be different from plaintext")

	// Decrypt
	decrypted, err := em.Decrypt(namespace, ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted, "decrypted data should match original")
}

func TestEncryptionManager_DisabledEncryption(t *testing.T) {
	// Set environment variable
	envKey := "test-encryption-key-123"
	err := os.Setenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY", envKey)
	assert.NoError(t, err)
	defer os.Unsetenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")

	// Create config with encryption disabled
	config := config.EncryptionConfig{
		Enabled: false,
	}

	// Create encryption manager
	em, err := NewEncryptionManager(config)
	assert.NoError(t, err)

	// Test data
	namespace := "test-namespace"
	plaintext := []byte("sensitive terraform state data")

	// Encrypt should return plaintext as-is
	ciphertext, err := em.Encrypt(namespace, plaintext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, ciphertext, "with encryption disabled, ciphertext should equal plaintext")

	// Decrypt should return ciphertext as-is
	decrypted, err := em.Decrypt(namespace, ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted, "with encryption disabled, decrypted should equal original")
}
