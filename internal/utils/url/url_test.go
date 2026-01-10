package url_test

import (
	"testing"

	"github.com/padok-team/burrito/internal/utils/url"
)

func TestNormalizeURLFullRepos(t *testing.T) {
	urlTypes := []string{
		"git@github.com:padok-team/burrito.git",
		"git@github.com:padok-team/burrito",
		"https://github.com/padok-team/burrito.git",
		"https://github.com/padok-team/burrito",
		"http://github.com/padok-team/burrito.git",
		"http://github.com/padok-team/burrito",
	}
	expected := "https://github.com/padok-team/burrito"
	for _, u := range urlTypes {
		normalized := url.NormalizeUrl(u)
		if normalized != expected {
			t.Errorf("Passed: %s, Expected %s, got %s", u, expected, normalized)
		}
	}
}

func TestNormalizeURLPrefixes(t *testing.T) {
	urlTypes := []string{
		"git@github.com:padok-team",
		"https://github.com/padok-team",
		"http://github.com/padok-team",
	}
	expected := "https://github.com/padok-team"
	for _, u := range urlTypes {
		normalized := url.NormalizeUrl(u)
		if normalized != expected {
			t.Errorf("Passed: %s, Expected %s, got %s", u, expected, normalized)
		}
	}
}

func TestNormalizeURLSSHWithCustomPort(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SSH URL without port",
			input:    "ssh://git@github.com/padok-team/burrito.git",
			expected: "https://github.com/padok-team/burrito",
		},
		{
			name:     "SSH URL with custom port 2222",
			input:    "ssh://git@example.com:2222/owner/repo.git",
			expected: "https://example.com/owner/repo",
		},
		{
			name:     "SSH URL with nested path and port",
			input:    "ssh://git@git.example.com:2222/group/subgroup/repo.git",
			expected: "https://git.example.com/group/subgroup/repo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalized := url.NormalizeUrl(tc.input)
			if normalized != tc.expected {
				t.Errorf("Input: %s, Expected: %s, Got: %s", tc.input, tc.expected, normalized)
			}
		})
	}
}

func TestNormalizeURLDomainOnly(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Domain only - github.com",
			input:    "github.com",
			expected: "https://github.com",
		},
		{
			name:     "Domain only - custom domain",
			input:    "git.example.com",
			expected: "https://git.example.com",
		},
		{
			name:     "Domain with git@ prefix",
			input:    "git@github.com",
			expected: "https://github.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalized := url.NormalizeUrl(tc.input)
			if normalized != tc.expected {
				t.Errorf("Input: %s, Expected: %s, Got: %s", tc.input, tc.expected, normalized)
			}
		})
	}
}

func TestNormalizeURLSelfHostedGitlab(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Self-hosted GitLab HTTPS with subpath",
			input:    "https://git.example.com/subpath/gitlab-manager",
			expected: "https://git.example.com/subpath/gitlab-manager",
		},
		{
			name:     "Self-hosted GitLab HTTPS with subpath and .git",
			input:    "https://git.example.com/subpath/gitlab-manager.git",
			expected: "https://git.example.com/subpath/gitlab-manager",
		},
		{
			name:     "Self-hosted GitLab SSH with subpath",
			input:    "git@git.example.com:subpath/gitlab-manager.git",
			expected: "https://git.example.com/subpath/gitlab-manager",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalized := url.NormalizeUrl(tc.input)
			if normalized != tc.expected {
				t.Errorf("Input: %s, Expected: %s, Got: %s", tc.input, tc.expected, normalized)
			}
		})
	}
}
