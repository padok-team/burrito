package comment

import "strings"

// The hidden marker lets providers update Burrito's previous PR/MR comment
// instead of creating duplicates when Kubernetes status persistence is retried.
const managedCommentMarker = "<!-- burrito:pull-request-comment -->"

type Comment interface {
	Generate(string) (string, error)
}

func WithManagedMarker(body string) string {
	if HasManagedMarker(body) {
		return body
	}
	return body + "\n\n" + managedCommentMarker
}

func HasManagedMarker(body string) bool {
	return strings.Contains(body, managedCommentMarker)
}
