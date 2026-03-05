package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// saveAndClearEnv saves the current values of the env vars used by the network mirror
// functions, clears them, and returns a cleanup function that restores the original values.
func saveAndClearEnv(t *testing.T) {
	t.Helper()
	keys := []string{"TG_PROVIDER_CACHE_DIR", "TF_PLUGIN_CACHE_DIR", "TF_CLI_CONFIG_FILE"}
	saved := make(map[string]string, len(keys))
	for _, k := range keys {
		saved[k] = os.Getenv(k)
		os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			os.Setenv(k, v)
		}
	})
}

func TestProviderCacheDir(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "no env vars set",
			envVars:  map[string]string{},
			expected: "",
		},
		{
			name:     "TG_PROVIDER_CACHE_DIR set",
			envVars:  map[string]string{"TG_PROVIDER_CACHE_DIR": "/tmp/tg-cache"},
			expected: "/tmp/tg-cache",
		},
		{
			name:     "TF_PLUGIN_CACHE_DIR set",
			envVars:  map[string]string{"TF_PLUGIN_CACHE_DIR": "/tmp/tf-cache"},
			expected: "/tmp/tf-cache",
		},
		{
			name: "both set, TG takes precedence",
			envVars: map[string]string{
				"TG_PROVIDER_CACHE_DIR": "/tmp/tg-cache",
				"TF_PLUGIN_CACHE_DIR":   "/tmp/tf-cache",
			},
			expected: "/tmp/tg-cache",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			saveAndClearEnv(t)
			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}

			got := providerCacheDir()
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestCreateNetworkMirrorConfig(t *testing.T) {
	t.Run("creates config with network mirror only", func(t *testing.T) {
		saveAndClearEnv(t)

		dir := t.TempDir()
		err := CreateNetworkMirrorConfig(dir, "https://mirror.example.com/")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "config.tfrc"))
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, `url = "https://mirror.example.com/"`) {
			t.Errorf("config missing network mirror URL, got:\n%s", contentStr)
		}
		if strings.Contains(contentStr, "plugin_cache_dir") {
			t.Errorf("config should not contain plugin_cache_dir, got:\n%s", contentStr)
		}

		if got := os.Getenv("TF_CLI_CONFIG_FILE"); got != filepath.Join(dir, "config.tfrc") {
			t.Errorf("TF_CLI_CONFIG_FILE expected %q, got %q", filepath.Join(dir, "config.tfrc"), got)
		}
	})

	t.Run("includes plugin_cache_dir when cache dir is set", func(t *testing.T) {
		saveAndClearEnv(t)
		os.Setenv("TF_PLUGIN_CACHE_DIR", "/tmp/provider-cache")

		dir := t.TempDir()
		err := CreateNetworkMirrorConfig(dir, "https://mirror.example.com/")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(dir, "config.tfrc"))
		if err != nil {
			t.Fatalf("failed to read config file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, `url = "https://mirror.example.com/"`) {
			t.Errorf("config missing network mirror URL, got:\n%s", contentStr)
		}
		if !strings.Contains(contentStr, `plugin_cache_dir = "/tmp/provider-cache"`) {
			t.Errorf("config missing plugin_cache_dir, got:\n%s", contentStr)
		}
		// plugin_cache_dir should appear before provider_installation
		cacheIdx := strings.Index(contentStr, "plugin_cache_dir")
		installIdx := strings.Index(contentStr, "provider_installation")
		if cacheIdx > installIdx {
			t.Errorf("plugin_cache_dir should appear before provider_installation in config")
		}
	})

	t.Run("returns error for invalid target path", func(t *testing.T) {
		saveAndClearEnv(t)

		err := CreateNetworkMirrorConfig("/nonexistent/path", "https://mirror.example.com/")
		if err == nil {
			t.Error("expected error for invalid path, got nil")
		}
	})
}
