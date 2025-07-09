package common

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

func TestReferenceName(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected plumbing.ReferenceName
	}{
		{
			name:     "branch name",
			ref:      "main",
			expected: plumbing.ReferenceName("refs/heads/main"),
		},
		{
			name:     "tag name",
			ref:      "v1.0.0",
			expected: plumbing.ReferenceName("refs/heads/v1.0.0"),
		},
		{
			name:     "full branch ref",
			ref:      "refs/heads/main",
			expected: plumbing.ReferenceName("refs/heads/main"),
		},
		{
			name:     "full tag ref",
			ref:      "refs/tags/v1.0.0",
			expected: plumbing.ReferenceName("refs/tags/v1.0.0"),
		},
		{
			name:     "remote branch ref",
			ref:      "refs/remotes/origin/main",
			expected: plumbing.ReferenceName("refs/remotes/origin/main"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReferenceName(tt.ref)
			if result != tt.expected {
				t.Errorf("ReferenceName(%q) = %q, want %q", tt.ref, result, tt.expected)
			}
		})
	}
}

func TestReferenceNameForTag(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected plumbing.ReferenceName
	}{
		{
			name:     "tag name",
			ref:      "v1.0.0",
			expected: plumbing.ReferenceName("refs/tags/v1.0.0"),
		},
		{
			name:     "branch name as tag",
			ref:      "main",
			expected: plumbing.ReferenceName("refs/tags/main"),
		},
		{
			name:     "full tag ref",
			ref:      "refs/tags/v1.0.0",
			expected: plumbing.ReferenceName("refs/tags/v1.0.0"),
		},
		{
			name:     "full branch ref",
			ref:      "refs/heads/main",
			expected: plumbing.ReferenceName("refs/heads/main"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReferenceNameForTag(tt.ref)
			if result != tt.expected {
				t.Errorf("ReferenceNameForTag(%q) = %q, want %q", tt.ref, result, tt.expected)
			}
		})
	}
}
