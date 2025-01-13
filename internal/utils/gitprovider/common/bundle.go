package common

import (
	"fmt"
	"os"
	"os/exec"
)

// Create a git bundle with `git bundle create` and return the content as a byte array
func CreateGitBundle(sourceDir, destination, ref string) ([]byte, error) {
	cmd := exec.Command("git", "-C", sourceDir, "bundle", "create", destination, ref)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create git bundle: %v, output: %s", err, string(output))
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to read git bundle: %v", err)
	}
	return data, nil
}
