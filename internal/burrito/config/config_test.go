package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a test configuration file
	configData := []byte(`
runner:
  action: "test"
controller:
  namespaces:
    - "ns1"
    - "ns2"
  timers:
    driftDetection: 10s
    onError: 5s
    waitAction: 1s
    failureGracePeriod: 30s
`)
	err := os.WriteFile("config.yaml", configData, 0644)
	if err != nil {
		t.Fatalf("failed to create test configuration file: %v", err)
	}
	defer os.Remove("config.yaml")

	// Create a test flag set
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
			Action: "test",
		},
		Controller: config.ControllerConfig{
			WatchedNamespaces: []string{"ns1", "ns2"},
			Timers: config.ControllerTimers{
				DriftDetection:     10 * time.Second,
				OnError:            5 * time.Second,
				WaitAction:         1 * time.Second,
				FailureGracePeriod: 30 * time.Second,
			},
		},
	}

	assert.Equal(t, expected, cfg)
}

// func TestLoadConfigWithEnvironmentVariables(t *testing.T) {
// 	// Set test environment variables
// 	err := os.Setenv("BURRITO_RUNNER_ACTION", "env-test")
// 	if err != nil {
// 		t.Fatalf("failed to set test environment variable: %v", err)
// 	}
// 	defer os.Unsetenv("BURRITO_RUNNER_ACTION")

// 	// Create a test flag set
// 	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
// 	flags.String("runner.action", "", "Runner action flag")

// 	// Create a test config instance
// 	cfg := &Config{}

// 	// Load the configuration
// 	err = cfg.Load(flags)
// 	if err != nil {
// 		t.Fatalf("failed to load configuration: %v", err)
// 	}

// 	// Assert the loaded values
// 	expected := &Config{
// 		Runner: RunnerConfig{
// 			Action: "env-test",
// 		},
// 	}

// 	if !reflect.DeepEqual(cfg, expected) {
// 		t.Errorf("loaded configuration does not match expected values.\nExpected: %+v\nActual: %+v", expected, cfg)
// 	}
// }
