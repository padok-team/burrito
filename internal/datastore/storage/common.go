package storage

import (
	"fmt"
	"strconv"

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
	Backend StorageBackend
	Config  config.Config
}

type StorageBackend interface {
	Get(key string) ([]byte, error)
	Check(key string) ([]byte, error)
	Set(key string, value []byte, ttl int) error
	Delete(key string) error
	List(prefix string) ([]string, error)
}

func New(config config.Config) Storage {
	switch {
	case config.Datastore.Storage.Azure.Container != "":
		return Storage{Backend: azure.New(config.Datastore.Storage.Azure, nil)}
	case config.Datastore.Storage.GCS.Bucket != "":
		return Storage{Backend: gcs.New(config.Datastore.Storage.GCS)}
	case config.Datastore.Storage.S3.Bucket != "":
		return Storage{Backend: s3.New(config.Datastore.Storage.S3)}
	case config.Datastore.Storage.Mock:
		log.Warn("Using mock storage backend - for testing only - data will only be stored in memory and will be lost when the process exits")
		return Storage{Backend: mock.New()}
	}
	return Storage{}
}

func (s *Storage) GetLogs(namespace string, layer string, run string, attempt string) ([]byte, error) {
	return s.Backend.Get(computeLogsKey(namespace, layer, run, attempt))
}

func (s *Storage) GetLatestLogs(namespace string, layer string, run string) ([]byte, error) {
	attempts, err := s.GetAttempts(namespace, layer, run)
	if err != nil {
		return nil, err
	}
	if attempts == 0 {
		return nil, &errors.StorageError{Nil: true}
	}
	attempt := strconv.Itoa(attempts - 1)
	return s.Backend.Get(computeLogsKey(namespace, layer, run, attempt))
}

func (s *Storage) PutLogs(namespace string, layer string, run string, attempt string, logs []byte) error {
	return s.Backend.Set(computeLogsKey(namespace, layer, run, attempt), logs, 0)
}

func (s *Storage) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	return s.Backend.Get(computePlanKey(namespace, layer, run, attempt, format))
}

func (s *Storage) GetLatestPlan(namespace string, layer string, run string, format string) ([]byte, error) {
	attempts, err := s.GetAttempts(namespace, layer, run)
	if err != nil {
		return nil, err
	}
	if attempts == 0 {
		return nil, &errors.StorageError{Nil: true}
	}
	attempt := strconv.Itoa(attempts - 1)
	return s.Backend.Get(computePlanKey(namespace, layer, run, attempt, format))
}

func (s *Storage) PutPlan(namespace string, layer string, run string, attempt string, format string, plan []byte) error {
	return s.Backend.Set(computePlanKey(namespace, layer, run, attempt, format), plan, 0)
}

func (s *Storage) GetAttempts(namespace string, layer string, run string) (int, error) {
	key := fmt.Sprintf("%s/%s/%s/%s", LayersPrefix, namespace, layer, run)
	attempts, err := s.Backend.List(key)
	return len(attempts), err
}

func (s *Storage) GetGitBundle(namespace string, repository string, ref string, commit string) ([]byte, error) {
	return s.Backend.Get(computeGitBundleKey(namespace, repository, ref, commit))
}

func (s *Storage) CheckGitBundle(namespace string, repository string, ref string, commit string) ([]byte, error) {
	return s.Backend.Check(computeGitBundleKey(namespace, repository, ref, commit))
}

func (s *Storage) PutGitBundle(namespace string, repository string, ref string, commit string, bundle []byte) error {
	// Store the git bundle
	err := s.Backend.Set(computeGitBundleKey(namespace, repository, ref, commit), bundle, 0)
	if err != nil {
		return fmt.Errorf("failed to store git bundle: %w", err)
	}
	return nil
}
