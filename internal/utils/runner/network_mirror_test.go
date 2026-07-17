package runner

import (
	"os"
	"testing"
)

func TestRemoveNetworkMirrorConfigMissingFile(t *testing.T) {
	_ = os.Setenv("TF_CLI_CONFIG_FILE", "stale-config")
	t.Cleanup(func() {
		_ = os.Unsetenv("TF_CLI_CONFIG_FILE")
	})

	if err := RemoveNetworkMirrorConfig(t.TempDir()); err != nil {
		t.Fatalf("RemoveNetworkMirrorConfig() error = %v", err)
	}
	if got := os.Getenv("TF_CLI_CONFIG_FILE"); got != "" {
		t.Fatalf("TF_CLI_CONFIG_FILE = %q, want empty", got)
	}
}
