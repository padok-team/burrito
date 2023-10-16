package url

import (
	"fmt"
	"strings"
)

// Normalize a Github/Gitlab URL (SSH or HTTPS) to a HTTPS URL
func NormalizeUrl(url string) string {
	if strings.Contains(url, "https://") {
		return removeGitExtension(url)
	}
	if strings.Contains(url, "http://") {
		return removeGitExtension("https://" + url[7:])
	}
	// All SSH URL from GitHub are like "git@padok.github.com:<owner>/<repo>.git"
	// We split on ":" then remove ".git" by removing the last characters
	// To handle enterprise GitHub, we dynamically get "padok.github.com"
	// By removing "git@" at the beginning of the string
	server, repo := splitSSHUrl(url)
	return fmt.Sprintf("https://%s/%s", server, repo)
}

func splitSSHUrl(url string) (server string, repo string) {
	split := strings.Split(url, ":")
	return split[0][4:], removeGitExtension(split[1])
}

func removeGitExtension(url string) string {
	if strings.HasSuffix(url, ".git") {
		return url[:len(url)-4]
	}
	return url
}
