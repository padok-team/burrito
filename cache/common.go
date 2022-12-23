package cache

import (
	"fmt"
	"hash/fnv"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

type CacheError struct {
	Err error
	Nil bool
}

func (c *CacheError) Error() string {
	return c.Err.Error()
}

func NotFound(err error) bool {
	ce, ok := err.(*CacheError)
	if ok {
		return ce.Nil
	}
	return false
}

func (c *CacheError) NotFound() bool {
	return c.Nil
}

type Prefix string

const (
	Lock                   Prefix = "lock"
	LastPlanDate           Prefix = "planDate"
	LastPlannedArtifact    Prefix = "plannedArtifact"
	LastAppliedArtifact    Prefix = "appliedArtifact"
	LastPlannedArtifactBin Prefix = "plannedArtifactBin"
	RunResult              Prefix = "runResult"
	RunMessage             Prefix = "runMessage"
)

type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl int) error
	Delete(key string) error
}

func GenerateKey(prefix Prefix, layer *configv1alpha1.TerraformLayer) string {
	var toHash string
	switch prefix {
	case Lock:
		toHash = layer.Spec.Repository.Name + layer.Spec.Repository.Namespace + layer.Spec.Path
		return fmt.Sprintf("%s-%d", prefix, hash(toHash))
	case LastPlanDate:
		toHash = layer.Spec.Repository.Name + layer.Spec.Repository.Namespace + layer.Spec.Path + layer.Spec.Branch
		return fmt.Sprintf("%s-%d", prefix, hash(toHash))
	case LastPlannedArtifact:
		toHash = layer.Spec.Repository.Name + layer.Spec.Repository.Namespace + layer.Spec.Path + layer.Spec.Branch
		return fmt.Sprintf("%s-%d", prefix, hash(toHash))
	case LastAppliedArtifact:
		toHash = layer.Spec.Repository.Name + layer.Spec.Repository.Namespace + layer.Spec.Path + layer.Spec.Branch
		return fmt.Sprintf("%s-%d", prefix, hash(toHash))
	case LastPlannedArtifactBin:
		toHash = layer.Spec.Repository.Name + layer.Spec.Repository.Namespace + layer.Spec.Path + layer.Spec.Branch
		return fmt.Sprintf("%s-%d", prefix, hash(toHash))
	case RunResult:
		toHash = layer.Spec.Repository.Name + layer.Spec.Repository.Namespace + layer.Spec.Path + layer.Spec.Branch
		return fmt.Sprintf("%s-%d", prefix, hash(toHash))
	case RunMessage:
		toHash = layer.Spec.Repository.Name + layer.Spec.Repository.Namespace + layer.Spec.Path + layer.Spec.Branch
		return fmt.Sprintf("%s-%d", prefix, hash(toHash))
	default:
		return ""
	}
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
