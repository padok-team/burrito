package gcs

import (
	"context"
	"fmt"
	"io"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"

	"github.com/padok-team/burrito/internal/burrito/config"

	gcs "cloud.google.com/go/storage"
)

type Storage struct {
	Bucket string
	Client *gcs.Client
}

func New(config config.GCS) (*Storage, error) {
	ctx := context.Background()
	client, err := gcs.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Storage{
		Bucket: config.Bucket,
		Client: client,
	}, nil
}

func (s *Storage) getFile(path string) ([]byte, error) {
	reader, err := s.Client.Bucket(s.Bucket).Object(path).NewReader(context.Background())
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	content := []byte{}
	_, err = reader.Read(content)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return content, nil
}

func (s *Storage) putFile(path string, content []byte) error {
	writer := s.Client.Bucket(s.Bucket).Object(path).NewWriter(context.Background())
	defer writer.Close()
	_, err := writer.Write(content)
	if err != nil {
		return err
	}
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
