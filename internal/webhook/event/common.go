package event

import (
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const PullRequestOpened = "opened"
const PullRequestClosed = "closed"

type ChangeInfo struct {
	ShaBefore string
	ShaAfter  string
}

type Event interface {
	Handle(client.Client) error
}

func ParseRevision(ref string) string {
	refParts := strings.SplitN(ref, "/", 3)
	return refParts[len(refParts)-1]
}

func isLayerLinkedToAnyRepositories(repositories []configv1alpha1.TerraformRepository, layer configv1alpha1.TerraformLayer) bool {
	for _, r := range repositories {
		if r.Name == layer.Spec.Repository.Name && r.Namespace == layer.Spec.Repository.Namespace {
			return true
		}
	}
	return false
}

func isPRLinkedToAnyRepositories(pr configv1alpha1.TerraformPullRequest, repos []configv1alpha1.TerraformRepository) bool {
	for _, r := range repos {
		if r.Name == pr.Spec.Repository.Name && r.Namespace == pr.Spec.Repository.Namespace {
			return true
		}
	}
	return false
}
