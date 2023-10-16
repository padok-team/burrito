package url_test

import (
	"testing"

	"github.com/padok-team/burrito/internal/utils/url"
)

func TestNormalizeURL(t *testing.T) {
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
