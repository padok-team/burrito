// Package commitstatus posts a single commit status scoped to one layer's plan/apply
// action. There is no pull request to aggregate multiple layers under a single status on
// a direct push to the base branch, so every affected layer gets its own status,
// disambiguated by its name.
package commitstatus

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v84/github"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
	logrus "github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

const (
	Needed     = "Pending"
	InProgress = "In progress"
	Succeeded  = "Success"
	Failed     = "Failure"

	// GitHub's Statuses API rejects descriptions over 140 characters with a 422.
	maxDescriptionLength = 140
)

// setStatusBackoff retries transient provider failures (e.g. a bare GitHub 500) up to 3
// times, with an increasing delay between attempts.
var setStatusBackoff = wait.Backoff{
	Duration: 200 * time.Millisecond,
	Factor:   2,
	Steps:    3,
}

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
	cs := status.CommitStatus{
		Phase:       phase,
		State:       state,
		Description: truncate(message, maxDescriptionLength),
		Commit:      commit,
		Context:     fmt.Sprintf("Burrito ▶ %s %s/%s", capitalize(string(phase)), layer.Namespace, layer.Name),
		TargetURL:   targetURL,
	}
	err := retry.OnError(setStatusBackoff, isRetryable, func() error {
		return provider.SetStatus(repository, nil, cs)
	})
	if err != nil {
		if isRetryable(err) {
			logrus.Warnf("could not set %s commit status for layer %s/%s after retries: %s", phase, layer.Namespace, layer.Name, err)
		} else {
			logrus.Warnf("commit status for layer %s/%s rejected by provider, not retrying: %s", layer.Namespace, layer.Name, err)
		}
	}
	return err
}

// isRetryable reports whether err looks transient (5xx, rate limiting, or a plain network
// failure with no HTTP response at all) and thus worth retrying, as opposed to a permanent
// client error (4xx, e.g. a validation failure or a provider-side cap being reached) that
// retrying identically would never fix.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response != nil {
		return ghErr.Response.StatusCode >= 500 || ghErr.Response.StatusCode == 429
	}
	var glErr *gitlab.ErrorResponse
	if errors.As(err, &glErr) && glErr.Response != nil {
		return glErr.Response.StatusCode >= 500 || glErr.Response.StatusCode == 429
	}
	// Unrecognized error shape (e.g. a network failure before any HTTP response) — assume
	// transient rather than silently dropping the status update.
	return true
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
