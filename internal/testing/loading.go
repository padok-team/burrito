package testing

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"

	"k8s.io/client-go/kubernetes/scheme"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// function that loads the contents of a folder into the cluster
func LoadResources(client client.Client, path string) {
	log := logf.FromContext(context.TODO())

	resources := parseResources(path)
	for _, r := range resources {
		log.Info(fmt.Sprintf("Creating %s, %s/%s", r.GetObjectKind().GroupVersionKind().Kind, r.GetNamespace(), r.GetName()))
		err := client.Create(context.TODO(), r)
		if err != nil {
			panic(err)
		}
	}
}

func parseResources(path string) []client.Object {
	log := logf.FromContext(context.TODO())
	_ = configv1alpha1.AddToScheme(scheme.Scheme)
	decoder := scheme.Codecs.UniversalDeserializer()

	list := []client.Object{}
	r := []byte{}
	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
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

	for _, doc := range strings.Split(string(r), "---") {
		if doc == "" || doc == "\n" {
			continue
		}
		obj, _, err := decoder.Decode([]byte(doc), nil, nil)
		if err != nil {
			log.Error(err, "Error while decoding YAML object")
			continue
		}
		list = append(list, obj.(client.Object))

	}
	return list
}
