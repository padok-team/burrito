package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type EncryptRequest struct {
	EncryptionKey string `json:"encryptionKey"`
}

type EncryptResponse struct {
	Message        string   `json:"message"`
	FilesEncrypted int      `json:"filesEncrypted"`
	Errors         []string `json:"errors,omitempty"`
}

func (a *API) EncryptAllFilesHandler(c echo.Context) error {
	// Parse the request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body",
		})
	}

	var req EncryptRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON format",
		})
	}

	if req.EncryptionKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "encryptionKey is required",
		})
	}

	// Get the configured encryption key from environment variable
	configuredKey := os.Getenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")
	if configuredKey == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "No encryption key configured on server",
		})
	}

	// Compare the provided key with the configured key
	if req.EncryptionKey != configuredKey {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid encryption key",
		})
	}

	// Check if encryption is enabled in configuration
	if !a.config.Datastore.Storage.Encryption.Enabled {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Encryption is not enabled in configuration",
		})
	}

	log.Info("Starting encryption of all files in datastore")

	// Encrypt all files
	filesEncrypted, errors := a.encryptAllFiles()

	response := EncryptResponse{
		Message:        fmt.Sprintf("Encryption process completed. %d files encrypted.", filesEncrypted),
		FilesEncrypted: filesEncrypted,
		Errors:         errors,
	}

	if len(errors) > 0 {
		log.Warnf("Encryption completed with %d errors", len(errors))
		return c.JSON(http.StatusPartialContent, response)
	}

	log.Infof("Successfully encrypted %d files", filesEncrypted)
	return c.JSON(http.StatusOK, response)
}

func (a *API) encryptAllFiles() (int, []string) {
	var filesEncrypted int
	var errors []string

	// List all files in both layers and repositories prefixes
	prefixes := []string{"layers"}

	for _, prefix := range prefixes {
		files, err := a.Storage.Backend.ListRecursive(prefix)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to list files in %s: %v", prefix, err))
			continue
		}

		for _, file := range files {
			// Skip if this looks like a directory (ends with /)
			if strings.HasSuffix(file, "/") {
				continue
			}

			err := a.encryptSingleFile(file)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to encrypt %s: %v", file, err))
				continue
			}

			filesEncrypted++
			if filesEncrypted%100 == 0 {
				log.Infof("Encrypted %d files so far...", filesEncrypted)
			}
		}
	}

	return filesEncrypted, errors
}

func (a *API) encryptSingleFile(filePath string) error {
	// Get the current file content
	data, err := a.Storage.Backend.Get(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Extract namespace from file path for encryption
	// File paths are typically: layers/namespace/layer/run/attempt/file or repositories/namespace/repo/branch/revision.gitbundle
	pathParts := strings.Split(strings.TrimPrefix(filePath, "/"), "/")
	if len(pathParts) < 2 {
		return fmt.Errorf("invalid file path format: %s", filePath)
	}

	namespace := pathParts[1] // Second part is always the namespace

	// Check if the file is already encrypted by trying to decrypt it directly with the encryptor
	encryptor := a.Storage.EncryptionManager.GetEncryptor(namespace)
	if encryptor == nil {
		return fmt.Errorf("no encryptor available for namespace %s", namespace)
	}

	// Try to decrypt the raw data directly. If it succeeds, the file is encrypted
	_, err = encryptor.Decrypt(data)
	if err == nil {
		// Decryption succeeded, so the file is already encrypted
		log.Infof("Skipping already encrypted file: %s", filePath)
		return nil
	}

	// Decryption failed, so the file is likely not encrypted. Encrypt it.
	encrypted, err := encryptor.Encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt: %w", err)
	}

	// Store the encrypted data back
	err = a.Storage.Backend.Set(filePath, encrypted, 0)
	if err != nil {
		return fmt.Errorf("failed to store encrypted file: %w", err)
	}

	log.Infof("Successfully encrypted file: %s", filePath)
	return nil
}
