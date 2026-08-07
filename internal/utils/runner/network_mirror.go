package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const networkMirrorConfigFile = "config.tfrc"

// Creates a network mirror configuration file for Terraform with the given endpoint
func CreateNetworkMirrorConfig(targetPath string, endpoint string) error {
	terraformrcContent := fmt.Sprintf(`
provider_installation {
  network_mirror {
   url = "%s"
  }
}`, endpoint)
	filePath := filepath.Join(targetPath, networkMirrorConfigFile)
	err := os.WriteFile(filePath, []byte(terraformrcContent), 0644)
	if err != nil {
		return err
	}
	err = os.Setenv("TF_CLI_CONFIG_FILE", filePath)
	if err != nil {
		return err
	}
	log.Infof("network mirror configuration created")
	return nil
}

// RemoveNetworkMirrorConfig removes the generated network mirror configuration file.
func RemoveNetworkMirrorConfig(targetPath string) error {
	filePath := filepath.Join(targetPath, networkMirrorConfigFile)
	err := os.Remove(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	err = os.Unsetenv("TF_CLI_CONFIG_FILE")
	if err != nil {
		return err
	}
	log.Infof("network mirror configuration removed")
	return nil
}
