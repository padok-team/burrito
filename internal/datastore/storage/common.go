package storage

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/storage/azure"
	errors "github.com/padok-team/burrito/internal/datastore/storage/error"
	"github.com/padok-team/burrito/internal/datastore/storage/gcs"
	"github.com/padok-team/burrito/internal/datastore/storage/mock"
	"github.com/padok-team/burrito/internal/datastore/storage/s3"
)

const (
	LogFile                string = "run.log"
	PlanBinFile            string = "plan.bin"
	PlanJsonFile           string = "plan.json"
	PrettyPlanFile         string = "pretty.plan"
	ShortDiffFile          string = "short.diff"
	GitBundleFileExtension string = ".gitbundle"
	RevisionFile           string = "latest"
	LayersPrefix           string = "layers"
	RepositoriesPrefix     string = "repositories"
)

func computeLogsKey(namespace string, layer string, run string, attempt string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s/%s", LayersPrefix, namespace, layer, run, attempt, LogFile)
}

func computePlanKey(namespace string, layer string, run string, attempt string, format string) string {
	key := ""
	prefix := fmt.Sprintf("%s/%s/%s/%s/%s", LayersPrefix, namespace, layer, run, attempt)
	switch format {
	case "json":
		key = fmt.Sprintf("%s/%s", prefix, PlanJsonFile)
	case "pretty":
		key = fmt.Sprintf("%s/%s", prefix, PrettyPlanFile)
	case "short":
		key = fmt.Sprintf("%s/%s", prefix, ShortDiffFile)
	case "bin":
		key = fmt.Sprintf("%s/%s", prefix, PlanBinFile)
	default:
		key = fmt.Sprintf("%s/%s", prefix, PlanJsonFile)
	}
	return key
}

func computeGitBundleKey(namespace string, repository string, branch string, revision string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s%s", RepositoriesPrefix, namespace, repository, branch, revision, GitBundleFileExtension)
}

type Storage struct {
	Backend           StorageBackend
	Config            config.Config
	EncryptionManager *EncryptionManager
}

type StorageBackend interface {
	Get(key string) ([]byte, error)
	Check(key string) ([]byte, error)
	Set(key string, value []byte, ttl int) error
	Delete(key string) error
	List(prefix string) ([]string, error)
	ListRecursive(prefix string) ([]string, error)
}

func New(config config.Config) Storage {
	em, err := NewEncryptionManager(config.Datastore.Storage.Encryption)
	if err != nil {
		log.Fatalf("Failed to create encryption manager: %v", err)
	}
	storage := Storage{
		Config:            config,
		EncryptionManager: em,
	}

	switch {
	case config.Datastore.Storage.Azure.Container != "":
		storage.Backend = azure.New(config.Datastore.Storage.Azure, nil)
	case config.Datastore.Storage.GCS.Bucket != "":
		storage.Backend = gcs.New(config.Datastore.Storage.GCS)
	case config.Datastore.Storage.S3.Bucket != "":
		storage.Backend = s3.New(config.Datastore.Storage.S3)
	case config.Datastore.Storage.Mock:
		log.Warn("Using mock storage backend - for testing only - data will only be stored in memory and will be lost when the process exits")
		storage.Backend = mock.New()
	}

	return storage
}

func (s *Storage) GetLogs(namespace string, layer string, run string, attempt string) ([]byte, error) {
	data, err := s.Backend.Get(computeLogsKey(namespace, layer, run, attempt))
	if err != nil {
		return nil, err
	} else {
		return s.EncryptionManager.Decrypt(namespace, data)
	}
}

func (s *Storage) GetLatestLogs(namespace string, layer string, run string) ([]byte, error) {
	latestAttempt, err := s.GetLatestAttempt(namespace, layer, run)
	if err != nil {
		return nil, err
	}
	if latestAttempt == "-1" {
		return nil, &errors.StorageError{Nil: true}
	}

	return s.GetLogs(namespace, layer, run, latestAttempt)
}

func (s *Storage) PutLogs(namespace string, layer string, run string, attempt string, logs []byte) error {
	dataToStore, err := s.EncryptionManager.Encrypt(namespace, logs)

	if err != nil {
		return err
	}

	err = s.Backend.Set(computeLogsKey(namespace, layer, run, attempt), dataToStore, 0)
	if err != nil {
		return fmt.Errorf("failed to store logs: %w", err)
	}
	return nil
}

func (s *Storage) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	data, err := s.Backend.Get(computePlanKey(namespace, layer, run, attempt, format))
	if err != nil {
		return nil, err
	} else {
		return s.EncryptionManager.Decrypt(namespace, data)
	}
}

func (s *Storage) GetLatestPlan(namespace string, layer string, run string, format string) ([]byte, error) {
	latestAttempt, err := s.GetLatestAttempt(namespace, layer, run)
	if err != nil {
		return nil, err
	}
	if latestAttempt == "-1" {
		return nil, &errors.StorageError{Nil: true}
	}

	return s.GetPlan(namespace, layer, run, latestAttempt, format)
}

func (s *Storage) PutPlan(namespace string, layer string, run string, attempt string, format string, plan []byte) error {
	dataToStore, err := s.EncryptionManager.Encrypt(namespace, plan)

	if err != nil {
		return err
	}

	err = s.Backend.Set(computePlanKey(namespace, layer, run, attempt, format), dataToStore, 0)
	if err != nil {
		return fmt.Errorf("failed to store plan: %w", err)
	}
	return nil
}

func (s *Storage) GetLatestAttempt(namespace string, layer string, run string) (string, error) {
	attempts, err := s.GetAttempts(namespace, layer, run)

	if err != nil || len(attempts) == 0 {
		return "-1", err
	}

	latestAttemptStr := attempts[len(attempts)-1]

	return latestAttemptStr, nil
}

func (s *Storage) GetAttempts(namespace string, layer string, run string) ([]string, error) {
	key := fmt.Sprintf("%s/%s/%s/%s", LayersPrefix, namespace, layer, run)
	attempts := []string{}
	paths, err := s.Backend.List(key)

	for _, path := range paths {
		// Remove the key prefix and trailing / to get just the attempt number
		// Example: /layers/ns/layer/run/0/ becomes 0,
		attemptStr := strings.TrimPrefix(path, "/"+key+"/")
		attemptId := strings.Split(attemptStr, "/")[0]

		attempts = append(attempts, attemptId)
	}

	// Sort the attempts numerically
	slices.SortFunc(attempts, func(a, b string) int {
		ai, _ := strconv.Atoi(a)
		bi, _ := strconv.Atoi(b)
		return ai - bi
	})

	return attempts, err
}

func (s *Storage) GetGitBundle(namespace string, repository string, ref string, commit string) ([]byte, error) {
	data, err := s.Backend.Get(computeGitBundleKey(namespace, repository, ref, commit))
	if err != nil {
		return nil, err
	} else {
		return s.EncryptionManager.Decrypt(namespace, data)
	}
}

func (s *Storage) CheckGitBundle(namespace string, repository string, ref string, commit string) ([]byte, error) {
	return s.Backend.Check(computeGitBundleKey(namespace, repository, ref, commit))
}

func (s *Storage) PutGitBundle(namespace string, repository string, ref string, commit string, bundle []byte) error {
	dataToStore, err := s.EncryptionManager.Encrypt(namespace, bundle)

	if err != nil {
		return err
	}

	// Store the git bundle
	err = s.Backend.Set(computeGitBundleKey(namespace, repository, ref, commit), dataToStore, 0)
	if err != nil {
		return fmt.Errorf("failed to store git bundle: %w", err)
	}
	return nil
}
