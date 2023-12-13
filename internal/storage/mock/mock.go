package mock

import (
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/storage"
)

type Storage struct {
	storage map[string][]byte
}

func New() *Storage {
	return &Storage{
		storage: make(map[string][]byte),
	}
}

func (s *Storage) getFile(path string) ([]byte, error) {
	value, ok := s.storage[path]
	if !ok {
		return nil, &storage.StorageError{
			Err: nil,
			Nil: false,
		}
	}
	return value, nil
}

func (s *Storage) putFile(path string, content []byte) error {
	s.storage[path] = content
	return nil
}

// Gets a file in the bucket in the below specified path:
// /logs/{layer}/{run}/{try}.log
func (s *Storage) GetLogs(run *configv1alpha1.TerraformRun) ([]byte, error) {
	path := fmt.Sprintf("/logs/%s-%s/%s-%s/%d.log", run.Spec.Layer.Namespace, run.Spec.Layer.Name, run.Namespace, run.Name, run.Status.Retries)
	return s.getFile(path)
}

// Gets a file in the bucket in the below specified path:
// /results/{layer}/{run}/plan.json
func (s *Storage) GetPlanArtifactJson(run *configv1alpha1.TerraformRun) ([]byte, error) {
	path := fmt.Sprintf("/results/%s-%s/%s-%s/plan.json", run.Spec.Layer.Namespace, run.Spec.Layer.Name, run.Namespace, run.Name)
	return s.getFile(path)
}

// Gets a file in the bucket in the below specified path:
// /results/{layer}/{run}/plan.bin
func (s *Storage) GetPlanArtifactBin(run *configv1alpha1.TerraformRun) ([]byte, error) {
	path := fmt.Sprintf("/results/%s-%s/%s-%s/plan.bin", run.Spec.Layer.Namespace, run.Spec.Layer.Name, run.Namespace, run.Name)
	return s.getFile(path)
}

// Gets a file in the bucket in the below specified path:
// /results/{layer}/{run}/plan.pretty
func (s *Storage) GetPrettyPlan(run *configv1alpha1.TerraformRun) ([]byte, error) {
	path := fmt.Sprintf("/results/%s-%s/%s-%s/plan.pretty", run.Spec.Layer.Namespace, run.Spec.Layer.Name, run.Namespace, run.Name)
	return s.getFile(path)
}

// Gets a file in the bucket in the below specified path:
// /git/{repository}/{branch}/{commit}.bundle
func (s *Storage) GetGitBundle(repository *configv1alpha1.TerraformRepository, commit string, branch string) ([]byte, error) {
	path := fmt.Sprintf("/git/%s-%s/%s/%s.bundle", repository.Namespace, repository.Name, branch, commit)
	return s.getFile(path)
}

// Puts a file in the bucket in the below specified path:
// /logs/{layer}/{run}/{try}.log
func (s *Storage) PutLogs(run *configv1alpha1.TerraformRun, logs []byte) error {
	path := fmt.Sprintf("/logs/%s-%s/%s-%s/%d.log", run.Spec.Layer.Namespace, run.Spec.Layer.Name, run.Namespace, run.Name, run.Status.Retries)
	return s.putFile(path, logs)
}

// Puts a file in the bucket in the below specified path:
// /results/{layer}/{run}/plan.json
func (s *Storage) PutPlanArtifactJson(run *configv1alpha1.TerraformRun, artifact []byte) error {
	path := fmt.Sprintf("/results/%s-%s/%s-%s/plan.json", run.Spec.Layer.Namespace, run.Spec.Layer.Name, run.Namespace, run.Name)
	return s.putFile(path, artifact)
}

// Puts a file in the bucket in the below specified path:
// /results/{layer}/{run}/plan.bin
func (s *Storage) PutPlanArtifactBin(run *configv1alpha1.TerraformRun, artifact []byte) error {
	path := fmt.Sprintf("/results/%s-%s/%s-%s/plan.bin", run.Spec.Layer.Namespace, run.Spec.Layer.Name, run.Namespace, run.Name)
	return s.putFile(path, artifact)
}

// Puts a file in the bucket in the below specified path:
// /results/{layer}/{run}/plan.pretty
func (s *Storage) PutPrettyPlan(run *configv1alpha1.TerraformRun, prettyPlan []byte) error {
	path := fmt.Sprintf("/results/%s-%s/%s-%s/plan.pretty", run.Spec.Layer.Namespace, run.Spec.Layer.Name, run.Namespace, run.Name)
	return s.putFile(path, prettyPlan)
}

// Puts a file in the bucket in the below specified path:
// /git/{repository}/{branch}/{commit}.bundle
func (s *Storage) PutGitBundle(repository *configv1alpha1.TerraformRepository, commit string, branch string, bundle []byte) error {
	path := fmt.Sprintf("/git/%s-%s/%s/%s.bundle", repository.Namespace, repository.Name, branch, commit)
	return s.putFile(path, bundle)
}
