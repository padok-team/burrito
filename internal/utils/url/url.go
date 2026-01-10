package url

import (
	"fmt"
	"net/url"
	"strings"
)

func NormalizeUrl(inputURL string) string {
	if strings.HasPrefix(inputURL, "https://") {
		return removeGitExtension(inputURL)
	}
	if strings.HasPrefix(inputURL, "http://") {
		return removeGitExtension("https://" + inputURL[7:])
	}
	if strings.HasPrefix(inputURL, "ssh://") {
		return normalizeSSHProtocolURL(inputURL)
	}
	return normalizeScpStyleURL(inputURL)
}

func normalizeSSHProtocolURL(inputURL string) string {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return normalizeScpStyleURL(inputURL)
	}
	server := parsedURL.Hostname()
	path := strings.TrimPrefix(parsedURL.Path, "/")
	return fmt.Sprintf("https://%s/%s", server, removeGitExtension(path))
}

func normalizeScpStyleURL(inputURL string) string {
	server, repo := splitScpStyleURL(inputURL)
	if repo == "" {
		return fmt.Sprintf("https://%s", server)
	}
	return fmt.Sprintf("https://%s/%s", server, repo)
}

func splitScpStyleURL(inputURL string) (server string, repo string) {
	colonIndex := strings.Index(inputURL, ":")
	if colonIndex == -1 {
		server = inputURL
		if strings.HasPrefix(server, "git@") {
			server = server[4:]
		}
		return server, ""
	}

	server = inputURL[:colonIndex]
	repo = inputURL[colonIndex+1:]

	if strings.HasPrefix(server, "git@") {
		server = server[4:]
	}

	return server, removeGitExtension(repo)
}

func removeGitExtension(inputURL string) string {
	if strings.HasSuffix(inputURL, ".git") {
		return inputURL[:len(inputURL)-4]
	}
	return inputURL
}
