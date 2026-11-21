package controllers

import (
	"testing"
	"time"

	"github.com/padok-team/burrito/internal/burrito"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/stretchr/testify/assert"
)

func TestControllersStartCmd_FlagOverrides(t *testing.T) {
	// Keep Load away from any configuration file in $HOME/.burrito/.
	t.Setenv("HOME", t.TempDir())

	app := &burrito.App{Config: &config.Config{}}
	cmd := buildControllersStartCmd(app)
	err := cmd.ParseFlags([]string{
		"--namespaces=ns1,ns2",
		"--types=layer",
		"--drift-detection-period=30m",
		"--terraform-max-retries=7",
		"--leader-election=false",
		"--log-format=json",
	})
	if err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}
	// Overridden by the environment, which takes precedence over flags.
	t.Setenv("BURRITO_CONTROLLER_MAXCONCURRENTRECONCILES", "4")

	if err := app.Config.Load(cmd.Flags()); err != nil {
		t.Fatalf("failed to load configuration: %v", err)
	}

	c := app.Config.Controller
	assert.Equal(t, []string{"ns1", "ns2"}, c.Namespaces)
	assert.Equal(t, []string{"layer"}, c.Types)
	assert.Equal(t, 30*time.Minute, c.Timers.DriftDetection)
	assert.Equal(t, 7, c.TerraformMaxRetries)
	assert.False(t, c.LeaderElection.Enabled)
	assert.Equal(t, "json", c.LogFormat)
	assert.Equal(t, 4, c.MaxConcurrentReconciles)
	// Flags left unset keep their defaults.
	assert.Equal(t, 10*time.Second, c.Timers.OnError)
	assert.Equal(t, ":8080", c.MetricsBindAddress)
}
