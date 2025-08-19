package storage

import (
	"fmt"
	"os"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/utils/encryption"
)

type EncryptionManager struct {
	DefaultEncryptor *encryption.Encryptor
	config           config.EncryptionConfig
}

func NewEncryptionManager(config config.EncryptionConfig) (*EncryptionManager, error) {
	em := &EncryptionManager{
		config: config,
	}

	encryptionKey := os.Getenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")

	if config.Enabled && encryptionKey == "" {
		return nil, fmt.Errorf("encryption is enabled but no encryption key is provided in the environment variable BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")
	} else if config.Enabled && encryptionKey != "" {
		encryptor, err := encryption.NewEncryptor(encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create encryptor: %w", err)
		}
		em.DefaultEncryptor = encryptor
	} else {
		em.DefaultEncryptor = nil
	}

	return em, nil
}

func (em *EncryptionManager) Encrypt(namespace string, plaintext []byte) ([]byte, error) {
	if em.DefaultEncryptor == nil {
		return plaintext, nil
	}

	return em.DefaultEncryptor.Encrypt(plaintext)
}

func (em *EncryptionManager) Decrypt(namespace string, ciphertext []byte) ([]byte, error) {
	if em.DefaultEncryptor == nil {
		return ciphertext, nil
	}

	// Try to decrypt the data. If it fails, return the original ciphertext as this might be a migration from an unencrypted state
	decrypted, err := em.DefaultEncryptor.Decrypt(ciphertext)
	if err != nil {
		return ciphertext, nil
	}

	return decrypted, nil
}
