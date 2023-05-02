package event

import (
	"fmt"
	"path/filepath"
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

// Normalize a Github/Gitlab URL (SSH or HTTPS) to a HTTPS URL
func NormalizeUrl(url string) string {
	if strings.Contains(url, "https://") {
		return url
	}
	if strings.Contains(url, "http://") {
		return "https://" + url[7:]
	}
	// All SSH URL from GitHub are like "git@padok.github.com:<owner>/<repo>.git"
	// We split on ":" then remove ".git" by removing the last characters
	// To handle enterprise GitHub, we dynamically get "padok.github.com"
	// By removing "git@" at the beginning of the string
	split := strings.Split(url, ":")
	return fmt.Sprintf("https://%s/%s", split[0][4:], split[1][:len(split[1])-4])
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

func layerFilesHaveChanged(layer configv1alpha1.TerraformLayer, changedFiles []string) bool {
	if len(changedFiles) == 0 {
		return true
	}

	// At last one changed file must be under refresh path
	for _, f := range changedFiles {
		f = ensureAbsPath(f)
		if strings.Contains(f, layer.Spec.Path) {
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

func ensureAbsPath(input string) string {
	if !filepath.IsAbs(input) {
		return string(filepath.Separator) + input
	}
	return input
}
