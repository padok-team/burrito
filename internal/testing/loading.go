package testing

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gruntwork-io/go-commons/errors"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	coordination "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Resources struct {
	Repositories []*configv1alpha1.TerraformRepository
	Layers       []*configv1alpha1.TerraformLayer
	Runs         []*configv1alpha1.TerraformRun
	PullRequests []*configv1alpha1.TerraformPullRequest
	Leases       []*coordination.Lease
}

type GenericResource struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec       interface{}       `json:"spec"`
}

func unmarshalYaml(data []byte, obj interface{}) error {
	jsonData, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		return errors.WithStackTrace(err)
	}
	err = json.Unmarshal(jsonData, obj)
	if err != nil {
		return errors.WithStackTrace(err)
	}
	return nil
}

// function that loads the contents of a folder into the cluster
func LoadResources(client client.Client, path string) {

	resources, err := parseResources(path)
	if err != nil {
		panic(err)
	}
	for _, r := range resources.Layers {
		deepCopy := r.DeepCopy()
		err := client.Create(context.TODO(), deepCopy)
		if err != nil {
			panic(err)
		}
		deepCopy.Status = r.Status
		err = client.Status().Update(context.TODO(), deepCopy)
		if err != nil {
			panic(err)
		}
	}
	for _, r := range resources.Repositories {
		deepCopy := r.DeepCopy()
		err := client.Create(context.TODO(), deepCopy)
		if err != nil {
			panic(err)
		}
		deepCopy.Status = r.Status
		err = client.Status().Update(context.TODO(), deepCopy)
		if err != nil {
			panic(err)
		}
	}
	for _, r := range resources.Runs {
		deepCopy := r.DeepCopy()
		err := client.Create(context.TODO(), deepCopy)
		if err != nil {
			panic(err)
		}
		deepCopy.Status = r.Status
		err = client.Status().Update(context.TODO(), deepCopy)
		if err != nil {
			panic(err)
		}
	}
	for _, r := range resources.Leases {
		deepCopy := r.DeepCopy()
		err := client.Create(context.TODO(), deepCopy)
		if err != nil {
			panic(err)
		}
	}
	for _, r := range resources.PullRequests {
		deepCopy := r.DeepCopy()
		err := client.Create(context.TODO(), deepCopy)
		if err != nil {
			panic(err)
		}
		deepCopy.Status = r.Status
		err = client.Status().Update(context.TODO(), deepCopy)
		if err != nil {
			panic(err)
		}
	}
}

func parseResources(path string) (*Resources, error) {
	log := logf.FromContext(context.TODO())
	_ = configv1alpha1.AddToScheme(scheme.Scheme)
	resources := Resources{}
	r := []byte{}
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, walkErr error) error {
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if string(data) == "" {
			return nil
		}
		r = append(r, []byte("\n---\n")...)
		r = append(r, data...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, doc := range strings.Split(string(r), "---") {
		if doc == "" || doc == "\n" {
			continue
		}
		genericResource := &GenericResource{}
		err := unmarshalYaml([]byte(doc), genericResource)
		if err != nil {
			log.Error(err, "Error while decoding YAML object")
			continue
		}
		switch genericResource.Kind {
		case "TerraformRepository":
			repo := &configv1alpha1.TerraformRepository{}
			err = yaml.Unmarshal([]byte(doc), &repo)
			resources.Repositories = append(resources.Repositories, repo)
		case "TerraformLayer":
			layer := &configv1alpha1.TerraformLayer{}
			err = yaml.Unmarshal([]byte(doc), &layer)
			resources.Layers = append(resources.Layers, layer)
		case "TerraformRun":
			run := &configv1alpha1.TerraformRun{}
			err = yaml.Unmarshal([]byte(doc), &run)
			resources.Runs = append(resources.Runs, run)
		case "Lease":
			lease := &coordination.Lease{}
			err = yaml.Unmarshal([]byte(doc), &lease)
			resources.Leases = append(resources.Leases, lease)
		case "TerraformPullRequest":
			pr := &configv1alpha1.TerraformPullRequest{}
			err = yaml.Unmarshal([]byte(doc), &pr)
			resources.PullRequests = append(resources.PullRequests, pr)
		default:
			continue
		}
		if err != nil {
			log.Error(err, "Error while decoding YAML object")
			continue
		}
	}
	return &resources, nil
}
