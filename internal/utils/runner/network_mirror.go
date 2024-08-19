package runner

import (
	"fmt"
	"os"
)

// Creates a network mirror configuration file for Terraform with the given endpoint
func CreateNetworkMirrorConfig(targetPath string, endpoint string) error {
	terraformrcContent := fmt.Sprintf(`
provider_installation {
  network_mirror {
   url = "%s"
  }
}`, endpoint)
	filePath := fmt.Sprintf("%s/config.tfrc", targetPath)
	err := os.WriteFile(filePath, []byte(terraformrcContent), 0644)
	if err != nil {
		return err
	}
	err = os.Setenv("TF_CLI_CONFIG_FILE", filePath)
	if err != nil {
		return err
	}
	return nil
}
