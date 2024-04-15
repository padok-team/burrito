package storage

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/storage/azure"
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

func New(config config.Config) *Storage {
	switch {
	case config.Datastore.Storage.Azure.Container != "":
		return &Storage{Backend: azure.New(config.Datastore.Storage.Azure)}
	case config.Datastore.Storage.GCS.Bucket != "":
		return &Storage{Backend: gcs.New(config.Datastore.Storage.GCS)}
	case config.Datastore.Storage.S3.Bucket != "":
		return &Storage{Backend: s3.New(config.Datastore.Storage.S3)}
	}
	return &Storage{}
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
	attempts, err := s.Backend.List(fmt.Sprintf("/%s/%s/%s", namespace, layer, run))
	if err != nil {
		return nil, err
	}
	if len(attempts) == 0 {
		return nil, &StorageError{Nil: true}
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

func (s *Storage) GetPlan(namespace string, layer string, run string, attempt string) ([]byte, error) {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, PlanBinFile)
	return s.Backend.Get(key)
}

func (s *Storage) PutPlan(namespace string, layer string, run string, attempt string, plan []byte) error {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, PlanBinFile)
	return s.Backend.Set(key, plan, int(s.Config.Controller.Timers.DriftDetection.Seconds()))
}

func (s *Storage) GetPlanJson(namespace string, layer string, run string, attempt string) ([]byte, error) {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, PlanJsonFile)
	return s.Backend.Get(key)
}

func (s *Storage) PutPlanJson(namespace string, layer string, run string, attempt string, plan []byte) error {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, PlanJsonFile)
	return s.Backend.Set(key, plan, int(s.Config.Controller.Timers.DriftDetection.Seconds()))
}

func (s *Storage) GetPrettyPlan(namespace string, layer string, run string, attempt string) ([]byte, error) {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, PrettyPlanFile)
	return s.Backend.Get(key)
}

func (s *Storage) PutPrettyPlan(namespace string, layer string, run string, attempt string, plan []byte) error {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, PrettyPlanFile)
	return s.Backend.Set(key, plan, int(s.Config.Controller.Timers.DriftDetection.Seconds()))
}

func (s *Storage) GetShortDiff(namespace string, layer string, run string, attempt string) ([]byte, error) {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, ShortDiffFile)
	return s.Backend.Get(key)
}

func (s *Storage) PutShortDiff(namespace string, layer string, run string, attempt string, diff []byte) error {
	key := fmt.Sprintf("/%s/%s/%s/%s/%s", namespace, layer, run, attempt, ShortDiffFile)
	return s.Backend.Set(key, diff, int(s.Config.Controller.Timers.DriftDetection.Seconds()))
}

type StorageError struct {
	Err error
	Nil bool
}

func (s *StorageError) Error() string {
	return s.Err.Error()
}

func NotFound(err error) bool {
	ce, ok := err.(*StorageError)
	if ok {
		return ce.Nil
	}
	return false
}

func (c *StorageError) NotFound() bool {
	return c.Nil
}

type Prefix string

const (
	LastPlannedArtifactBin  Prefix = "plannedArtifactBin"
	RunMessage              Prefix = "runMessage"
	LastPlannedArtifactJson Prefix = "plannedArtifactJson"
	LastPlanResult          Prefix = "planResult"
	LastPrettyPlan          Prefix = "prettyPlan"
)

type StorageBackend interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl int) error
	Delete(key string) error
	List(prefix string) ([]string, error)
}

func GenerateKey(prefix Prefix, layer *configv1alpha1.TerraformLayer) string {
	toHash := layer.Spec.Repository.Name + layer.Spec.Repository.Namespace + layer.Spec.Path + layer.Spec.Branch
	return fmt.Sprintf("%s-%d", prefix, hash(toHash))
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
