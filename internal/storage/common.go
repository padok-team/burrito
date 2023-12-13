package storage

import (
	"fmt"
	"hash/fnv"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

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

type Storage interface {
	GetLogs(run *configv1alpha1.TerraformRun) ([]byte, error)
	GetPlanArtifactJson(run *configv1alpha1.TerraformRun) ([]byte, error)
	GetPlanArtifactBin(run *configv1alpha1.TerraformRun) ([]byte, error)
	GetPrettyPlan(run *configv1alpha1.TerraformRun) ([]byte, error)
	GetGitBundle(repository *configv1alpha1.TerraformRepository, commit string, branch string) ([]byte, error)
	PutLogs(run *configv1alpha1.TerraformRun, logs []byte) error
	PutPlanArtifactJson(run *configv1alpha1.TerraformRun, artifact []byte) error
	PutPlanArtifactBin(run *configv1alpha1.TerraformRun, artifact []byte) error
	PutPrettyPlan(run *configv1alpha1.TerraformRun, prettyPlan []byte) error
	PutGitBundle(repository *configv1alpha1.TerraformRepository, commit string, branch string, bundle []byte) error
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
