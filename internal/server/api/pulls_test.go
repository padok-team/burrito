package api

import "testing"

func TestPullRequestURL(t *testing.T) {
	testCases := []struct {
		name     string
		repoURL  string
		id       string
		expected string
	}{
		{
			name:     "GitHub HTTPS with .git",
			repoURL:  "https://github.com/padok-team/burrito.git",
			id:       "42",
			expected: "https://github.com/padok-team/burrito/pull/42",
		},
		{
			name:     "GitHub HTTPS without .git",
			repoURL:  "https://github.com/padok-team/burrito",
			id:       "42",
			expected: "https://github.com/padok-team/burrito/pull/42",
		},
		{
			name:     "GitHub HTTP",
			repoURL:  "http://github.com/padok-team/burrito.git",
			id:       "42",
			expected: "https://github.com/padok-team/burrito/pull/42",
		},
		{
			name:     "GitHub SCP-style SSH",
			repoURL:  "git@github.com:padok-team/burrito.git",
			id:       "42",
			expected: "https://github.com/padok-team/burrito/pull/42",
		},
		{
			name:     "GitHub ssh:// with custom port",
			repoURL:  "ssh://git@github.com:2222/padok-team/burrito.git",
			id:       "42",
			expected: "https://github.com/padok-team/burrito/pull/42",
		},
		{
			name:     "GitLab HTTPS",
			repoURL:  "https://gitlab.com/padok-team/burrito.git",
			id:       "7",
			expected: "https://gitlab.com/padok-team/burrito/-/merge_requests/7",
		},
		{
			name:     "GitLab SCP-style SSH",
			repoURL:  "git@gitlab.com:padok-team/burrito.git",
			id:       "7",
			expected: "https://gitlab.com/padok-team/burrito/-/merge_requests/7",
		},
		{
			name:     "GitLab nested groups",
			repoURL:  "https://gitlab.com/group/subgroup/burrito.git",
			id:       "7",
			expected: "https://gitlab.com/group/subgroup/burrito/-/merge_requests/7",
		},
		{
			name:     "Self-hosted GitLab detected from the host",
			repoURL:  "https://gitlab.example.com/padok-team/burrito",
			id:       "7",
			expected: "https://gitlab.example.com/padok-team/burrito/-/merge_requests/7",
		},
		{
			name:     "GitLab host detection is case-insensitive",
			repoURL:  "https://GitLab.example.com/padok-team/burrito",
			id:       "7",
			expected: "https://GitLab.example.com/padok-team/burrito/-/merge_requests/7",
		},
		{
			name:     "gitlab in the path but not in the host is not GitLab",
			repoURL:  "https://git.example.com/subpath/gitlab-manager.git",
			id:       "3",
			expected: "https://git.example.com/subpath/gitlab-manager/pull/3",
		},
		{
			name:     "Unknown host defaults to GitHub-style",
			repoURL:  "https://git.example.com/padok-team/burrito",
			id:       "3",
			expected: "https://git.example.com/padok-team/burrito/pull/3",
		},
		{
			name:     "Empty repository URL",
			repoURL:  "",
			id:       "3",
			expected: "",
		},
		{
			name:     "Empty pull request ID",
			repoURL:  "https://github.com/padok-team/burrito",
			id:       "",
			expected: "",
		},
		{
			name:     "Unparsable repository URL",
			repoURL:  "https://github.com:bad port/padok-team/burrito",
			id:       "3",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := pullRequestURL(tc.repoURL, tc.id)
			if got != tc.expected {
				t.Errorf("repoURL: %s, id: %s, expected: %q, got: %q", tc.repoURL, tc.id, tc.expected, got)
			}
		})
	}
}
