package encryption

import (
	"bytes"
	"testing"
)

func TestEncryptor_EncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		plaintext []byte
	}{
		{
			name:      "small text",
			key:       "test-key-123",
			plaintext: []byte("hello world"),
		},
		{
			name:      "terraform plan json",
			key:       "terraform-plan-key",
			plaintext: []byte(`{"format_version":"1.1","terraform_version":"1.0.0","planned_values":{}}`),
		},
		{
			name:      "large binary data",
			key:       "binary-key",
			plaintext: make([]byte, 10000), // 10KB of zeros
		},
		{
			name:      "empty data",
			key:       "empty-key",
			plaintext: []byte{},
		},
		{
			name:      "unicode text",
			key:       "unicode-key",
			plaintext: []byte("Hello 世界 🌍"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor, err := NewEncryptor(tt.key)
			if err != nil {
				t.Fatalf("failed to create encryptor: %v", err)
			}

			// Encrypt the plaintext
			ciphertext, err := encryptor.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("encryption failed: %v", err)
			}

			// For non-empty plaintext, ensure ciphertext is different
			if len(tt.plaintext) > 0 && bytes.Equal(tt.plaintext, ciphertext) {
				t.Error("ciphertext should be different from plaintext")
			}

			// Decrypt the ciphertext
			decrypted, err := encryptor.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("decryption failed: %v", err)
			}

			// Verify the decrypted data matches the original
			if !bytes.Equal(tt.plaintext, decrypted) {
				t.Errorf("decrypted data doesn't match original\noriginal: %v\ndecrypted: %v", tt.plaintext, decrypted)
			}
		})
	}
}

func TestEncryptor_DifferentKeys(t *testing.T) {
	plaintext := []byte("sensitive terraform plan data")

	encryptor1, err := NewEncryptor("key1")
	if err != nil {
		t.Fatalf("failed to create encryptor1: %v", err)
	}
	encryptor2, err := NewEncryptor("key2")
	if err != nil {
		t.Fatalf("failed to create encryptor2: %v", err)
	}

	// Encrypt with first key
	ciphertext, err := encryptor1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Try to decrypt with second key (should fail)
	_, err = encryptor2.Decrypt(ciphertext)
	if err == nil {
		t.Error("decryption should fail with different key")
	}
}

func TestEncryptor_KeyHashing(t *testing.T) {
	// Test that different length keys work correctly due to hashing
	keys := []string{
		"short",
		"this-is-a-medium-length-key",
		"this-is-a-very-long-key-that-exceeds-32-bytes-and-should-be-hashed-properly",
	}

	plaintext := []byte("test data for key hashing")

	for _, key := range keys {
		t.Run("key_length_"+string(rune(len(key))), func(t *testing.T) {
			encryptor, err := NewEncryptor(key)
			if err != nil {
				t.Fatalf("failed to create encryptor: %v", err)
			}

			ciphertext, err := encryptor.Encrypt(plaintext)
			if err != nil {
				t.Fatalf("encryption failed: %v", err)
			}

			decrypted, err := encryptor.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("decryption failed: %v", err)
			}

			if !bytes.Equal(plaintext, decrypted) {
				t.Error("decrypted data doesn't match original")
			}
		})
	}
}

func TestEncryptor_NonceUniqueness(t *testing.T) {
	encryptor, err := NewEncryptor("test-key")
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}
	plaintext := []byte("same plaintext")

	// Encrypt the same plaintext multiple times
	ciphertexts := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		ciphertext, err := encryptor.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("encryption %d failed: %v", i, err)
		}
		ciphertexts[i] = ciphertext
	}

	// Verify all ciphertexts are different (due to random nonces)
	for i := 0; i < len(ciphertexts); i++ {
		for j := i + 1; j < len(ciphertexts); j++ {
			if bytes.Equal(ciphertexts[i], ciphertexts[j]) {
				t.Errorf("ciphertexts %d and %d should be different", i, j)
			}
		}
	}
}

func TestEncryptor_InvalidCiphertext(t *testing.T) {
	encryptor, err := NewEncryptor("test-key")
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{
			name:       "too short",
			ciphertext: []byte("short"),
		},
		{
			name:       "corrupted data",
			ciphertext: make([]byte, 100), // 100 bytes of zeros
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.ciphertext)
			if err == nil {
				t.Error("decryption should fail for invalid ciphertext")
			}
		})
	}
}

func BenchmarkEncryptor_Encrypt(b *testing.B) {
	encryptor, err := NewEncryptor("benchmark-key")
	if err != nil {
		b.Fatalf("failed to create encryptor: %v", err)
	}
	plaintext := make([]byte, 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Encrypt(plaintext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncryptor_Decrypt(b *testing.B) {
	encryptor, err := NewEncryptor("benchmark-key")
	if err != nil {
		b.Fatalf("failed to create encryptor: %v", err)
	}
	plaintext := make([]byte, 1024) // 1KB

	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Decrypt(ciphertext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestNewEncryptor_EmptyKey(t *testing.T) {
	_, err := NewEncryptor("")
	if err == nil {
		t.Error("NewEncryptor should return an error for empty key")
	}

	expectedError := "encryption key cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}
}
