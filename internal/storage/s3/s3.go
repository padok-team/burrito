package s3

import (
	"bytes"
	"fmt"
	"io"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/storage"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	s3 "github.com/aws/aws-sdk-go/service/s3"
)

type Storage struct {
	Bucket string
	Client *s3.S3
}

func New(config config.S3) (*Storage, error) {
	session := session.Must(session.NewSession())
	client := s3.New(session)
	return &Storage{
		Client: client,
		Bucket: config.Bucket,
	}, nil
}

func (s *Storage) getFile(path string) ([]byte, error) {
	resp, err := s.Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return nil, &storage.StorageError{
					Err: err,
					Nil: true,
				}
			}
		}
		return nil, err
	}
	defer resp.Body.Close()
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Storage) putFile(path string, content []byte) error {
	_, err := s.Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(content),
	})
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
