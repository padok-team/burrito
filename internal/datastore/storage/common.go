package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/storage/azure"
	errors "github.com/padok-team/burrito/internal/datastore/storage/error"
	"github.com/padok-team/burrito/internal/datastore/storage/gcs"
	"github.com/padok-team/burrito/internal/datastore/storage/s3"
)

const (
	LogFile        string = "run.log"
	PlanBinFile    string = "plan.bin"
	PlanJsonFile   string = "plan.json"
	PrettyPlanFile string = "pretty.plan"
	ShortDiffFile  string = "short.diff"
)

type Storage struct {
	Backend StorageBackend
	Config  config.Config
}

type StorageBackend interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl int) error
	Delete(key string) error
	List(prefix string) ([]string, error)
}

func New(config config.Config) Storage {
	switch {
	case config.Datastore.Storage.Azure.Container != "":
		return Storage{Backend: azure.New(config.Datastore.Storage.Azure)}
	case config.Datastore.Storage.GCS.Bucket != "":
		return Storage{Backend: gcs.New(config.Datastore.Storage.GCS)}
	case config.Datastore.Storage.S3.Bucket != "":
		return Storage{Backend: s3.New(config.Datastore.Storage.S3)}
	}
	return Storage{}
}

func last(a []string) string {
	return a[len(a)-1]
}

func getMax(l []string) (int, error) {
	max := 0
	for _, v := range l {
		value, err := strconv.Atoi(last(strings.Split(v, "/")))
		if err != nil {
			return 0, err
		}
		if value > max {
			max = value
		}
	}
	return max, nil
}

func (s *Storage) GetLogs(namespace string, layer string, run string, attempt string) ([]byte, error) {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, LogFile)
	return s.Backend.Get(key)
}

func (s *Storage) GetLatestLogs(namespace string, layer string, run string) ([]byte, error) {
	attempts, err := s.Backend.List(fmt.Sprintf("/%s/%s/%s/", namespace, layer, run))
	if err != nil {
		return nil, err
	}
	if len(attempts) == 0 {
		return nil, &errors.StorageError{Nil: true}
	}
	attempt, err := getMax(attempts)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("/%s/%s/%s/%d/%s", namespace, layer, run, attempt, LogFile)
	return s.Backend.Get(key)
}

func (s *Storage) PutLogs(namespace string, layer string, run string, attempt string, logs []byte) error {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, LogFile)
	return s.Backend.Set(key, logs, 0)
}

func computePlanKey(namespace string, layer string, run string, attempt string, format string) string {
	key := ""
	prefix := fmt.Sprintf("/%s/%s/%s/%s", namespace, layer, run, attempt)
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

func (s *Storage) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	key := computePlanKey(namespace, layer, run, attempt, format)
	return s.Backend.Get(key)
}

func (s *Storage) GetLatestPlan(namespace string, layer string, run string, format string) ([]byte, error) {
	attempts, err := s.Backend.List(fmt.Sprintf("/%s/%s/%s/", namespace, layer, run))
	if err != nil {
		return nil, err
	}
	if len(attempts) == 0 {
		return nil, &errors.StorageError{Nil: true}
	}
	attempt, err := getMax(attempts)
	if err != nil {
		return nil, err
	}
	key := computePlanKey(namespace, layer, run, strconv.Itoa(attempt), format)
	return s.Backend.Get(key)
}

func (s *Storage) PutPlan(namespace string, layer string, run string, attempt string, format string, plan []byte) error {
	key := computePlanKey(namespace, layer, run, attempt, format)
	return s.Backend.Set(key, plan, 0)
}
