// Package commitstatus posts a single commit status scoped to one layer's plan/apply
// action. There is no pull request to aggregate multiple layers under a single status on
// a direct push to the base branch, so every affected layer gets its own status,
// disambiguated by its name.
package commitstatus

import (
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
)

const (
	Needed     = "Pending"
	InProgress = "In progress"
	Succeeded  = "Success"
	Failed     = "Failure"

	// GitHub's Statuses API rejects descriptions over 140 characters with a 422.
	maxDescriptionLength = 140
)

// Post sets a commit status for layer's phase (plan or apply), on the given commit. The
// pending/success/failure state is carried by GitHub/GitLab's own native status field, shown
// next to the context — so message (which mirrors the layer's "Last Result" field: the short
// plan/apply summary, or a placeholder like "Layer has never been planned") doesn't need to
// repeat it. It is truncated as needed since it's the only part with unbounded length.
// targetURL, if not empty, becomes the status's "Details" link (see LogsURL).
//
// No emoji prefix here: GitHub's Statuses API rejects any 4-byte UTF-8 character (which
// covers almost every modern emoji, e.g. 🌯) in the description with a 422.
func Post(provider repositorytypes.APIProvider, repository *configv1alpha1.TerraformRepository, layer *configv1alpha1.TerraformLayer, phase status.Phase, state status.State, commit string, message string, targetURL string) error {
	return provider.SetStatus(repository, nil, status.CommitStatus{
		Phase:       phase,
		State:       state,
		Description: truncate(message, maxDescriptionLength),
		Commit:      commit,
		Context:     fmt.Sprintf("Burrito ▶ %s %s/%s", capitalize(string(phase)), layer.Namespace, layer.Name),
		TargetURL:   targetURL,
	})
}

// LogsURL builds a link to layer's logs page on the Burrito dashboard at publicURL, scoped
// to runName when known. Returns "" when publicURL is unset (links are opt-in), leaving the
// status without a "Details" link.
func LogsURL(publicURL string, layer *configv1alpha1.TerraformLayer, runName string) string {
	if publicURL == "" {
		return ""
	}
	url := fmt.Sprintf("%s/logs/%s/%s", strings.TrimRight(publicURL, "/"), layer.Namespace, layer.Name)
	if runName != "" {
		url += "/" + runName
	}
	return url
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}
