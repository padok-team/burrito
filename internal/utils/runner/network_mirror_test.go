package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveNetworkMirrorConfigMissingFile(t *testing.T) {
	_ = os.Setenv("TF_CLI_CONFIG_FILE", "stale-config")
	t.Cleanup(func() {
		_ = os.Unsetenv("TF_CLI_CONFIG_FILE")
	})

	if err := RemoveNetworkMirrorConfig(t.TempDir(), "http://hermitcrab.local"); err != nil {
		t.Fatalf("RemoveNetworkMirrorConfig() error = %v", err)
	}
	if got := os.Getenv("TF_CLI_CONFIG_FILE"); got != "" {
		t.Fatalf("TF_CLI_CONFIG_FILE = %q, want empty", got)
	}
}

func TestRemoveNetworkMirrorConfigPreservesCustomSettings(t *testing.T) {
	t.Setenv("TF_CLI_CONFIG_FILE", t.TempDir()+"/config.tfrc")
	if err := os.WriteFile(os.Getenv("TF_CLI_CONFIG_FILE"), []byte("custom settings"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := RemoveNetworkMirrorConfig(filepath.Dir(os.Getenv("TF_CLI_CONFIG_FILE")), "http://hermitcrab.local"); err == nil {
		t.Fatal("expected refusal to remove custom settings")
	}
	if content, err := os.ReadFile(os.Getenv("TF_CLI_CONFIG_FILE")); err != nil || string(content) != "custom settings" {
		t.Fatalf("custom settings were lost: %q, %v", content, err)
	}
}
