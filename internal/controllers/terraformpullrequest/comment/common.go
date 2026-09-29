package comment

import "strings"

// The hidden marker lets providers update Burrito's previous PR/MR comment
// instead of creating duplicates when Kubernetes status persistence is retried.
const managedCommentMarker = "<!-- burrito:pull-request-comment -->"

type Comment interface {
	Generate(string) (string, error)
	// Marker identifies the comment this instance owns on the PR/MR.
	Marker() string
}

// Marker returns the hidden marker for an instance. An empty instance keeps the
// historical marker so comments posted before instances were named are still found.
func Marker(instance string) string {
	if instance == "" {
		return managedCommentMarker
	}
	return "<!-- burrito:pull-request-comment:" + instance + " -->"
}

func WithManagedMarker(body, marker string) string {
	if HasManagedMarker(body, marker) {
		return body
	}
	return body + "\n\n" + marker
}

func HasManagedMarker(body, marker string) bool {
	return strings.Contains(body, marker)
}
