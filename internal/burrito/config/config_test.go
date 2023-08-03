package config_test

import (
	"os"
	"testing"
	"time"

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
			Repository: config.RepositoryConfig{
				SSHPrivateKey: "private-key",
				Username:      "test",
				Password:      "password",
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
			},
			Types: []string{"layer", "repository", "pullrequest"},
			LeaderElection: config.LeaderElectionConfig{
				Enabled: true,
				ID:      "6d185457.terraform.padok.cloud",
			},
			MetricsBindAddress:     ":8080",
			HealthProbeBindAddress: ":8081",
			KubernetesWebhookPort:  9443,
			GithubConfig: config.GithubConfig{
				APIToken: "github-token",
			},
			GitlabConfig: config.GitlabConfig{
				APIToken: "gitlab-token",
				URL:      "https://gitlab.example.com",
			},
		},
		Redis: config.Redis{
			Hostname:   "burrito-redis.namespace",
			ServerPort: 5000,
			Database:   0,
			Password:   "testPassword",
		},
		Server: config.ServerConfig{
			Addr: ":9090",
			Webhook: config.WebhookConfig{
				Github: config.WebhookGithubConfig{
					Secret: "github-secret",
				},
				Gitlab: config.WebhookGitlabConfig{
					Secret: "gitlab-secret",
				},
			},
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
	setEnvVar(t, "BURRITO_RUNNER_REPOSITORY_USERNAME", "other-username", &envVarList)
	setEnvVar(t, "BURRITO_RUNNER_REPOSITORY_PASSWORD", "other-password", &envVarList)
	setEnvVar(t, "BURRITO_RUNNER_REPOSITORY_SSHPRIVATEKEY", "other-private-key", &envVarList)
	// Redis
	setEnvVar(t, "BURRITO_REDIS_HOSTNAME", "other-redis", &envVarList)
	setEnvVar(t, "BURRITO_REDIS_SERVERPORT", "8000", &envVarList)
	setEnvVar(t, "BURRITO_REDIS_DATABASE", "1", &envVarList)
	setEnvVar(t, "BURRITO_REDIS_PASSWORD", "otherPassword", &envVarList)
	// Controller
	setEnvVar(t, "BURRITO_CONTROLLER_TYPES", "layer,repository", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_NAMESPACES", "default,burrito,other", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TIMERS_DRIFTDETECTION", "10m", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TIMERS_ONERROR", "30s", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TIMERS_WAITACTION", "30s", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_TIMERS_FAILUREGRACEPERIOD", "1m", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_LEADERELECTION_ID", "other-leader-id", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_GITHUBCONFIG_APITOKEN", "pr-github-token", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_GITLABCONFIG_APITOKEN", "mr-gitlab-token", &envVarList)
	setEnvVar(t, "BURRITO_CONTROLLER_GITLABCONFIG_URL", "https://gitlab.com", &envVarList)
	// Server
	setEnvVar(t, "BURRITO_SERVER_ADDR", ":8090", &envVarList)
	setEnvVar(t, "BURRITO_SERVER_WEBHOOK_GITHUB_SECRET", "other-github-secret", &envVarList)
	setEnvVar(t, "BURRITO_SERVER_WEBHOOK_GITLAB_SECRET", "other-gitlab-secret", &envVarList)

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
			Repository: config.RepositoryConfig{
				SSHPrivateKey: "other-private-key",
				Username:      "other-username",
				Password:      "other-password",
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
			},
			Types: []string{"layer", "repository"},
			LeaderElection: config.LeaderElectionConfig{
				Enabled: true,
				ID:      "other-leader-id",
			},
			MetricsBindAddress:     ":8080",
			HealthProbeBindAddress: ":8081",
			KubernetesWebhookPort:  9443,
			GithubConfig: config.GithubConfig{
				APIToken: "pr-github-token",
			},
			GitlabConfig: config.GitlabConfig{
				APIToken: "mr-gitlab-token",
				URL:      "https://gitlab.com",
			},
		},
		Redis: config.Redis{
			Hostname:   "other-redis",
			ServerPort: 8000,
			Database:   1,
			Password:   "otherPassword",
		},
		Server: config.ServerConfig{
			Addr: ":8090",
			Webhook: config.WebhookConfig{
				Github: config.WebhookGithubConfig{
					Secret: "other-github-secret",
				},
				Gitlab: config.WebhookGitlabConfig{
					Secret: "other-gitlab-secret",
				},
			},
		},
	}

	assert.Equal(t, expected, cfg)
}

// TODO: make that test pass
// func TestConfig_FlagOverrides(t *testing.T) {
// 	// Read the test configuration file
// 	configFile, err := os.ReadFile("testdata/test-config-1.yaml")
// 	if err != nil {
// 		t.Fatalf("failed to read test configuration file: %v", err)
// 	}

// 	err = os.WriteFile("config.yaml", configFile, 0644)
// 	if err != nil {
// 		t.Fatalf("failed to create test configuration file: %v", err)
// 	}
// 	defer os.Remove("config.yaml")

// 	// Create an empty test flag set
// 	flags := pflag.NewFlagSet("controllers", pflag.ContinueOnError)
// 	flags.String("drift-detection-period", "1m", "drift-detection-period")

// 	// Create a test config instance
// 	cfg := &config.Config{}

// 	// Load the configuration
// 	err = cfg.Load(flags)
// 	if err != nil {
// 		t.Fatalf("failed to load configuration: %v", err)
// 	}

// 	// Assert the loaded values
// 	expected := &config.Config{
// 		Runner: config.RunnerConfig{
// 			Action: "apply",
// 			Layer: config.Layer{
// 				Name:      "test",
// 				Namespace: "default",
// 			},
// 			Repository: config.RepositoryConfig{
// 				SSHPrivateKey: "private-key",
// 				Username:      "test",
// 				Password:      "password",
// 			},
// 			SSHKnownHostsConfigMapName: "burrito-ssh-known-hosts",
// 		},
// 		Controller: config.ControllerConfig{
// 			Namespaces: []string{"default", "burrito"},
// 			Timers: config.ControllerTimers{
// 				DriftDetection:     1 * time.Minute,
// 				OnError:            1 * time.Minute,
// 				WaitAction:         1 * time.Minute,
// 				FailureGracePeriod: 15 * time.Second,
// 			},
// 			Types: []string{"layer", "repository", "pullrequest"},
// 			LeaderElection: config.LeaderElectionConfig{
// 				Enabled: true,
// 				ID:      "6d185457.terraform.padok.cloud",
// 			},
// 			MetricsBindAddress:     ":8080",
// 			HealthProbeBindAddress: ":8081",
// 			KubernetesWebhookPort:  9443,
// 			GithubConfig: config.GithubConfig{
// 				APIToken: "github-token",
// 			},
// 			GitlabConfig: config.GitlabConfig{
// 				APIToken: "gitlab-token",
// 				URL:      "https://gitlab.example.com",
// 			},
// 		},
// 		Redis: config.Redis{
// 			URL:      "burrito-redis:6379",
// 			Database: 0,
// 			Password: "testPassword",
// 		},
// 		Server: config.ServerConfig{
// 			Addr: ":8080",
// 			Webhook: config.WebhookConfig{
// 				Github: config.WebhookGithubConfig{
// 					Secret: "github-secret",
// 				},
// 				Gitlab: config.WebhookGitlabConfig{
// 					Secret: "gitlab-secret",
// 				},
// 			},
// 		},
// 	}

// 	assert.Equal(t, expected, cfg)
// }
