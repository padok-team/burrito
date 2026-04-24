package comment

import (
	"errors"
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeDatastore struct {
	plans       map[string][]byte
	err         error
	errByFormat map[string]error
}

func (f *fakeDatastore) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	if err, ok := f.errByFormat[format]; ok {
		return nil, err
	}
	return f.plans[format], nil
}

func (f *fakeDatastore) PutPlan(namespace string, layer string, run string, attempt string, format string, content []byte) error {
	return nil
}

func (f *fakeDatastore) GetLogs(namespace string, layer string, run string, attempt string) ([]string, error) {
	return nil, nil
}

func (f *fakeDatastore) PutLogs(namespace string, layer string, run string, attempt string, content []byte) error {
	return nil
}

func (f *fakeDatastore) PutGitBundle(namespace string, name string, ref string, revision string, bundle []byte) error {
	return nil
}

func (f *fakeDatastore) CheckGitBundle(namespace string, name string, ref string, revision string) (bool, error) {
	return false, nil
}

func (f *fakeDatastore) GetGitBundle(namespace string, name string, ref string, revision string) ([]byte, error) {
	return nil, nil
}

func TestDefaultCommentGenerate(t *testing.T) {
	comment := NewDefaultComment([]configv1alpha1.TerraformLayer{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "layer-a",
				Namespace: "default",
			},
			Spec: configv1alpha1.TerraformLayerSpec{
				Path: "terraform/",
			},
			Status: configv1alpha1.TerraformLayerStatus{
				LastRun: configv1alpha1.TerraformLayerRun{
					Name: "run-a",
				},
			},
		},
	}, &fakeDatastore{
		plans: map[string][]byte{
			"pretty": []byte("pretty plan"),
			"short":  []byte("+ create"),
		},
	})

	got, err := comment.Generate("abc123")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	for _, expected := range []string{"Burrito Report", "layer-a", "terraform/", "+ create", "pretty plan", "abc123"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected generated comment to contain %q, got:\n%s", expected, got)
		}
	}
}

func TestDefaultCommentGenerateReturnsDatastoreError(t *testing.T) {
	expectedErr := errors.New("datastore unavailable")
	comment := NewDefaultComment([]configv1alpha1.TerraformLayer{{}}, &fakeDatastore{err: expectedErr})

	_, err := comment.Generate("abc123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestDefaultCommentGenerateReturnsShortPlanError(t *testing.T) {
	expectedErr := errors.New("short plan unavailable")
	comment := NewDefaultComment([]configv1alpha1.TerraformLayer{{}}, &fakeDatastore{
		plans: map[string][]byte{
			"pretty": []byte("pretty plan"),
		},
		errByFormat: map[string]error{
			"short": expectedErr,
		},
	})

	_, err := comment.Generate("abc123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestManagedMarkerHelpers(t *testing.T) {
	body := "hello"
	withMarker := WithManagedMarker(body)
	if !HasManagedMarker(withMarker) {
		t.Fatalf("expected generated body to contain managed marker")
	}
	if got := WithManagedMarker(withMarker); got != withMarker {
		t.Fatalf("expected WithManagedMarker to be idempotent")
	}
	if HasManagedMarker(body) {
		t.Fatalf("did not expect unmanaged body to contain marker")
	}
}

func TestNewInitialComment(t *testing.T) {
	if NewInitialComment() == nil {
		t.Fatalf("expected NewInitialComment to return a comment")
	}
}
