package runner

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// providerCacheDir returns the filesystem path to use as a local provider cache,
// detected from environment variables. Returns "" if no cache directory is configured.
func providerCacheDir() string {
	// TG_PROVIDER_CACHE_DIR is set by Terragrunt users to share providers across modules.
	if dir := os.Getenv("TG_PROVIDER_CACHE_DIR"); dir != "" {
		return dir
	}
	// TF_PLUGIN_CACHE_DIR is the standard OpenTofu/Terraform provider cache.
	if dir := os.Getenv("TF_PLUGIN_CACHE_DIR"); dir != "" {
		return dir
	}
	return ""
}

// CreateNetworkMirrorConfig creates a provider_installation configuration file that
// routes provider downloads through the given network mirror endpoint (hermitcrab).
//
// When a local provider cache directory is detected (via TG_PROVIDER_CACHE_DIR or
// TF_PLUGIN_CACHE_DIR), a plugin_cache_dir directive is added so that providers are
// cached locally alongside the network mirror. This is necessary because an explicit
// provider_installation block causes both OpenTofu and Terraform to ignore the
// TF_PLUGIN_CACHE_DIR environment variable entirely.
func CreateNetworkMirrorConfig(targetPath string, endpoint string) error {
	var pluginCacheLine string
	if cacheDir := providerCacheDir(); cacheDir != "" {
		pluginCacheLine = fmt.Sprintf("plugin_cache_dir = \"%s\"\n", cacheDir)
		log.Infof("detected provider cache directory: %s", cacheDir)
	}

	terraformrcContent := fmt.Sprintf(`%sprovider_installation {
  network_mirror {
    url = "%s"
  }
}
`, pluginCacheLine, endpoint)

	filePath := fmt.Sprintf("%s/config.tfrc", targetPath)
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
