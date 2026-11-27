package config_test

import (
	"os"
	"testing"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestConfig_FromYamlFile(t *testing.T) {
	// Read the test configuration file
	configFile, err := os.ReadFile("testdata/test-config-1.yaml")
	if err != nil {
		t.Fatalf("failed to read test configuration file: %v", err)
	}

	err = os.WriteFile("config.yaml", configFile, 0644)
	if err != nil {
		t.Fatalf("failed to create test configuration file: %v", err)
	}
	defer os.Remove("config.yaml")

	// Create an empty test flag set
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	// flags.String("runner.action", "", "Runner action flag")

	// Create a test config instance
	cfg := &config.Config{}

	// Load the configuration
	err = cfg.Load(flags)
	if err != nil {
		t.Fatalf("failed to load configuration: %v", err)
	}

	// Assert the loaded values
	expected := &config.Config{
		Runner: config.RunnerConfig{
			Action: "apply",
			Layer: config.Layer{
				Name:      "test",
				Namespace: "default",
			},
			Image: config.ImageConfig{
				Repository: "test-repository",
				Tag:        "test-tag",
				PullPolicy: "Always",
			},
			SSHKnownHostsConfigMapName: "burrito-ssh-known-hosts",
		},
		Controller: config.ControllerConfig{
			Namespaces: []string{"default", "burrito"},
			Timers: config.ControllerTimers{
				DriftDetection:     20 * time.Minute,
				OnError:            1 * time.Minute,
				WaitAction:         1 * time.Minute,
				FailureGracePeriod: 15 * time.Second,
				CredentialsTTL:     1 * time.Hour,
			},
			DefaultSyncWindows: []configv1alpha1.SyncWindow{
				{
					Kind:     configv1alpha1.SyncWindowKindDeny,
					Schedule: "0 0 * * *",
					Duration: "1h",
					Layers:   []string{"layer1", "layer2"},
					Actions:  []string{"plan", "apply"},
				},
			},
			TerraformMaxRetries:     5,
			MaxConcurrentReconciles: 1,
			MaxConcurrentRunnerPods: 0,
			CommitStatus:            config.CommitStatusConfig{Enabled: true},
			Types:                   []string{"layer", "repository", "run", "pullrequest"},
			LeaderElection: config.LeaderElectionConfig{
				Enabled: true,
				ID:      "6d185457.terraform.padok.cloud",
			},
			MetricsBindAddress:     ":8080",
			HealthProbeBindAddress: ":8081",
			KubernetesWebhookPort:  9443,
			LogFormat:              "text",
		},
		Datastore: config.DatastoreConfig{
			LogFormat: "text",
		},
		Server: config.ServerConfig{
			Addr: ":9090",
		},
	}

	assert.Equal(t, expected, cfg)
}

func setEnvVar(t *testing.T, key, value string, list *[]string) {
	err := os.Setenv(key, value)
	if err != nil {
		t.Fatalf("failed to set test environment variable: %v", err)
	}
	*list = append(*list, key)
}

func TestConfig_EnvVarOverrides(t *testing.T) {
	// Read the test configuration file
	configFile, err := os.ReadFile("testdata/test-config-1.yaml")
	if err != nil {
		t.Fatalf("failed to read test configuration file: %v", err)
	}

	err = os.WriteFile("config.yaml", configFile, 0644)
	if err != nil {
		t.Fatalf("failed to create test configuration file: %v", err)
	}
	defer os.Remove("config.yaml")

	// Create an empty test flag set
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	// flags.String("drift-detection", "30m", "drift detection period flag")

	// Set environment variables
	envVarList := []string{}
	// Runner
	setEnvVar(t, "BURRITO_RUNNER_ACTION", "plan", &envVarList)
	setEnvVar(t, "BURRITO_RUNNER_LAYER_NAME", "other-layer", &envVarList)
	setEnvVar(t, "BURRITO_RUNNER_LAYER_NAMESPACE", "other-namespace", &envVarList)
	setEnvVar(t, "BURRITO_RUNNER_IMAGE_REPOSITORY", "other-repository", &envVarList)
	setEnvVar(t, "BURRITO_RUNNER_IMAGE_TAG", "other-tag", &envVarList)
	setEnvVar(t, "BURRITO_RUNNER_IMAGE_PULLPOLICY", "Always", &envVarList)
	// Controller
	setEnvVar(t, "BURRITO_CONTROLLER_TYPES", "layer,repository", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_NAMESPACES", "default,burrito,other", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TIMERS_DRIFTDETECTION", "10m", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TIMERS_ONERROR", "30s", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TIMERS_WAITACTION", "30s", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TIMERS_FAILUREGRACEPERIOD", "1m", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TIMERS_CREDENTIALSTTL", "2h", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_MAXCONCURRENTRECONCILES", "3", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_MAXCONCURRENTRUNNERPODS", "10", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TERRAFORMMAXRETRIES", "32", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_LEADERELECTION_ID", "other-leader-id", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_LOGFORMAT", "json", &envVarList)
	// Datastore
	setEnvVar(t, "BURRITO_DATASTORE_LOGFORMAT", "json", &envVarList)
	// Server
	setEnvVar(t, "BURRITO_SERVER_ADDR", ":8090", &envVarList)

	// Create a test config instance
	cfg := &config.Config{}

	// Load the configuration
	err = cfg.Load(flags)
	if err != nil {
		t.Fatalf("failed to load configuration: %v", err)
	}

	// Unset the environment variables
	for _, envVar := range envVarList {
		err = os.Unsetenv(envVar)
		if err != nil {
			t.Fatalf("failed to unset test environment variable: %v", err)
		}
	}

	// Assert the loaded values
	expected := &config.Config{
		Runner: config.RunnerConfig{
			Action: "plan",
			Layer: config.Layer{
				Name:      "other-layer",
				Namespace: "other-namespace",
			},
			Image: config.ImageConfig{
				Repository: "other-repository",
				Tag:        "other-tag",
				PullPolicy: "Always",
			},
			SSHKnownHostsConfigMapName: "burrito-ssh-known-hosts",
		},
		Controller: config.ControllerConfig{
			Namespaces: []string{"default", "burrito", "other"},
			Timers: config.ControllerTimers{
				DriftDetection:     10 * time.Minute,
				OnError:            30 * time.Second,
				WaitAction:         30 * time.Second,
				FailureGracePeriod: 1 * time.Minute,
				CredentialsTTL:     2 * time.Hour,
			},
			DefaultSyncWindows: []configv1alpha1.SyncWindow{
				{
					Kind:     configv1alpha1.SyncWindowKindDeny,
					Schedule: "0 0 * * *",
					Duration: "1h",
					Layers:   []string{"layer1", "layer2"},
					Actions:  []string{"plan", "apply"},
				},
			},
			MaxConcurrentReconciles: 3,
			MaxConcurrentRunnerPods: 10,
			CommitStatus:            config.CommitStatusConfig{Enabled: true},
			TerraformMaxRetries:     32,
			Types:                   []string{"layer", "repository"},
			LeaderElection: config.LeaderElectionConfig{
				Enabled: true,
				ID:      "other-leader-id",
			},
			MetricsBindAddress:     ":8080",
			HealthProbeBindAddress: ":8081",
			KubernetesWebhookPort:  9443,
			LogFormat:              "json",
		},
		Datastore: config.DatastoreConfig{
			LogFormat: "json",
		},
		Server: config.ServerConfig{
			Addr: ":8090",
		},
	}

	assert.Equal(t, expected, cfg)
}

// writeTestConfigFile copies testdata/test-config-1.yaml to ./config.yaml, where Load looks
// for it, and removes it when the test ends.
func writeTestConfigFile(t *testing.T) {
	t.Helper()
	configFile, err := os.ReadFile("testdata/test-config-1.yaml")
	if err != nil {
		t.Fatalf("failed to read test configuration file: %v", err)
	}
	err = os.WriteFile("config.yaml", configFile, 0644)
	if err != nil {
		t.Fatalf("failed to create test configuration file: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove("config.yaml"); err != nil {
			t.Errorf("failed to remove test configuration file: %v", err)
		}
	})
}

// newControllerFlags mimics how cmd/ declares flags: each flag writes straight into the
// Config field it targets, before Load runs. Load then unmarshals the configuration file
// and environment variables on top of the struct.
func newControllerFlags(cfg *config.Config) *pflag.FlagSet {
	flags := pflag.NewFlagSet("controllers", pflag.ContinueOnError)
	flags.DurationVar(&cfg.Controller.Timers.DriftDetection, "drift-detection-period", 4*time.Hour, "")
	flags.DurationVar(&cfg.Controller.Timers.RepositorySync, "repository-sync-period", 5*time.Minute, "")
	flags.IntVar(&cfg.Controller.TerraformMaxRetries, "terraform-max-retries", 5, "")
	flags.StringVar(&cfg.Controller.LogFormat, "log-format", "text", "")
	return flags
}

func loadWithArgs(t *testing.T, args []string) *config.Config {
	t.Helper()
	cfg := &config.Config{}
	flags := newControllerFlags(cfg)
	if err := flags.Parse(args); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}
	if err := cfg.Load(flags); err != nil {
		t.Fatalf("failed to load configuration: %v", err)
	}
	return cfg
}

func TestConfig_FlagOverrides(t *testing.T) {
	t.Run("flag default is kept when neither the config file nor env set the key", func(t *testing.T) {
		cfg := loadWithArgs(t, nil)
		assert.Equal(t, 4*time.Hour, cfg.Controller.Timers.DriftDetection)
		assert.Equal(t, 5*time.Minute, cfg.Controller.Timers.RepositorySync)
		assert.Equal(t, "text", cfg.Controller.LogFormat)
	})

	t.Run("explicit flag is used when neither the config file nor env set the key", func(t *testing.T) {
		writeTestConfigFile(t)
		// timers.repositorySync is absent from the configuration file.
		cfg := loadWithArgs(t, []string{"--repository-sync-period=7m"})
		assert.Equal(t, 7*time.Minute, cfg.Controller.Timers.RepositorySync)
		// Keys the flags don't target still come from the configuration file.
		assert.Equal(t, []string{"default", "burrito"}, cfg.Controller.Namespaces)
	})

	t.Run("explicit flag is used when env sets an unrelated key", func(t *testing.T) {
		t.Setenv("BURRITO_CONTROLLER_LOGFORMAT", "json")
		cfg := loadWithArgs(t, []string{"--terraform-max-retries=9"})
		assert.Equal(t, 9, cfg.Controller.TerraformMaxRetries)
		assert.Equal(t, "json", cfg.Controller.LogFormat)
	})
}

// TestConfig_Precedence pins the current order: environment variable > configuration file >
// CLI flag (explicit or default). Flags are applied to the struct before Load unmarshals the
// file and the environment on top, so any key set in the file wins over an explicit flag.
// Note that this differs from the order stated in #144/#146 (flags above the file).
func TestConfig_Precedence(t *testing.T) {
	t.Run("config file overrides an explicit flag", func(t *testing.T) {
		writeTestConfigFile(t)
		// timers.driftDetection is 20m in the configuration file.
		cfg := loadWithArgs(t, []string{"--drift-detection-period=1m"})
		assert.Equal(t, 20*time.Minute, cfg.Controller.Timers.DriftDetection)
	})

	t.Run("env overrides an explicit flag", func(t *testing.T) {
		t.Setenv("BURRITO_CONTROLLER_TERRAFORMMAXRETRIES", "32")
		cfg := loadWithArgs(t, []string{"--terraform-max-retries=9"})
		assert.Equal(t, 32, cfg.Controller.TerraformMaxRetries)
	})

	t.Run("env overrides both the config file and an explicit flag", func(t *testing.T) {
		writeTestConfigFile(t)
		t.Setenv("BURRITO_CONTROLLER_TIMERS_DRIFTDETECTION", "10m")
		cfg := loadWithArgs(t, []string{"--drift-detection-period=1m"})
		assert.Equal(t, 10*time.Minute, cfg.Controller.Timers.DriftDetection)
	})
}
