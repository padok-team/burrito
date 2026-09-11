package commitstatus

import (
	"errors"
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeAPIProvider struct {
	setStatusCalls []status.CommitStatus
	// setStatusErrs, if set, is consumed one error per call (nil entries succeed);
	// once exhausted, further calls succeed.
	setStatusErrs []error
}

func (p *fakeAPIProvider) GetChanges(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest) ([]string, error) {
	return nil, nil
}

func (p *fakeAPIProvider) Comment(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, c comment.Comment) error {
	return nil
}

func (p *fakeAPIProvider) ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error) {
	return nil, nil
}

func (p *fakeAPIProvider) SetStatus(repository *configv1alpha1.TerraformRepository, pullRequest *configv1alpha1.TerraformPullRequest, s status.CommitStatus) error {
	p.setStatusCalls = append(p.setStatusCalls, s)
	if len(p.setStatusErrs) > 0 {
		err := p.setStatusErrs[0]
		p.setStatusErrs = p.setStatusErrs[1:]
		return err
	}
	return nil
}

func TestPostTruncatesLongDescriptionToGitHubLimit(t *testing.T) {
	provider := &fakeAPIProvider{}
	repository := &configv1alpha1.TerraformRepository{ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"}}
	layer := &configv1alpha1.TerraformLayer{ObjectMeta: metav1.ObjectMeta{Name: "pwet", Namespace: "default"}}
	longMessage := strings.Repeat("Plan: 1 to add, 0 to change, 0 to destroy. ", 10)

	err := Post(provider, repository, layer, status.PhasePlan, status.StateSuccess, "sha123", longMessage, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(provider.setStatusCalls) != 1 {
		t.Fatalf("expected exactly one commit status to be set, got %d", len(provider.setStatusCalls))
	}
	got := provider.setStatusCalls[0].Description
	if len([]rune(got)) > maxDescriptionLength {
		t.Fatalf("expected description to be at most %d runes, got %d: %q", maxDescriptionLength, len([]rune(got)), got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected truncated description to end with an ellipsis, got %q", got)
	}
}

func TestPostKeepsShortDescriptionUntouched(t *testing.T) {
	provider := &fakeAPIProvider{}
	repository := &configv1alpha1.TerraformRepository{ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"}}
	layer := &configv1alpha1.TerraformLayer{ObjectMeta: metav1.ObjectMeta{Name: "pwet", Namespace: "default"}}

	err := Post(provider, repository, layer, status.PhaseApply, status.StateFailure, "sha123", "short message", "https://burrito.example.com/logs/default/pwet/run-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	call := provider.setStatusCalls[0]
	if call.Description != "short message" {
		t.Errorf("expected description %q, got %q", "short message", call.Description)
	}
	wantContext := "Burrito ▶ Apply default/pwet"
	if call.Context != wantContext {
		t.Errorf("expected context %q, got %q", wantContext, call.Context)
	}
	wantTargetURL := "https://burrito.example.com/logs/default/pwet/run-1"
	if call.TargetURL != wantTargetURL {
		t.Errorf("expected target URL %q, got %q", wantTargetURL, call.TargetURL)
	}
}

func TestLogsURL(t *testing.T) {
	layer := &configv1alpha1.TerraformLayer{ObjectMeta: metav1.ObjectMeta{Name: "pwet", Namespace: "default"}}

	if got := LogsURL("", layer, "run-1"); got != "" {
		t.Errorf("expected an empty URL when publicURL is unset, got %q", got)
	}
	if got, want := LogsURL("https://burrito.example.com", layer, ""), "https://burrito.example.com/logs/default/pwet?hidepr=false"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
	if got, want := LogsURL("https://burrito.example.com/", layer, "run-1"), "https://burrito.example.com/logs/default/pwet/run-1?hidepr=false"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestPostReturnsTheProviderErrorWithoutRetrying(t *testing.T) {
	providerErr := errors.New("500 Internal Server Error")
	provider := &fakeAPIProvider{setStatusErrs: []error{providerErr, nil}}
	repository := &configv1alpha1.TerraformRepository{ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"}}
	layer := &configv1alpha1.TerraformLayer{ObjectMeta: metav1.ObjectMeta{Name: "pwet", Namespace: "default"}}

	err := Post(provider, repository, layer, status.PhasePlan, status.StateSuccess, "sha123", "message", "")
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected the provider error to be returned, got %v", err)
	}
	if len(provider.setStatusCalls) != 1 {
		t.Fatalf("expected exactly 1 attempt, got %d", len(provider.setStatusCalls))
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"hello", 5, "hello"},
	}
	for _, c := range cases {
		if got := truncate(c.in, c.max); got != c.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}

func TestStableContext(t *testing.T) {
	short := "Burrito ▶ Apply default/pwet"
	long := "Burrito ▶ Plan " + strings.Repeat("very-long-namespace-name-", 20) + "/" + strings.Repeat("very-long-layer-name-", 20)

	// Short context unchanged
	if got := stableContext(short); got != short {
		t.Errorf("short context changed: got %q, want %q", got, short)
	}

	// Long context truncated to exactly 255 runes
	got := stableContext(long)
	if len([]rune(got)) != 255 {
		t.Errorf("long context truncated to %d runes, want 255", len([]rune(got)))
	}

	// Determinism: same input always produces same output
	got1 := stableContext(long)
	got2 := stableContext(long)
	if got1 != got2 {
		t.Errorf("non-deterministic output: %q != %q", got1, got2)
	}

	// Two contexts sharing prefix get different results due to hash suffix
	prefix := "Burrito ▶ Plan " + strings.Repeat("x", 240)
	ctx1 := prefix + "AAA"
	ctx2 := prefix + "BBB"
	res1 := stableContext(ctx1)
	res2 := stableContext(ctx2)
	if res1 == res2 {
		t.Errorf("different inputs with same prefix produced identical truncated contexts: %q", res1)
	}
}
