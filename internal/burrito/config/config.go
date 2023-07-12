package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Runner     RunnerConfig     `yaml:"runner"`
	Controller ControllerConfig `yaml:"controller"`
	Redis      Redis            `yaml:"redis"`
	Server     Server           `yaml:"server"`
}

type WebhookConfig struct {
	Github WebhookGithubConfig `yaml:"github"`
	Gitlab WebhookGitlabConfig `yaml:"gitlab"`
}

type WebhookGithubConfig struct {
	Secret string `yaml:"secret"`
}

type WebhookGitlabConfig struct {
	URL      string `yaml:"url"`
	Secret   string `yaml:"secret"`
	APIToken string `yaml:"token"`
}

type ControllerConfig struct {
	WatchedNamespaces      []string             `yaml:"namespaces"`
	Timers                 ControllerTimers     `yaml:"timers"`
	Types                  []string             `yaml:"types"`
	LeaderElection         LeaderElectionConfig `yaml:"leaderElection"`
	MetricsBindAddress     string               `yaml:"metricsBindAddress"`
	HealthProbeBindAddress string               `yaml:"healthProbeBindAddress"`
	KubernetesWehbookPort  int                  `yaml:"kubernetesWebhookPort"`
	GithubConfig           GithubConfig         `yaml:"githubConfig"`
	GitlabConfig           GitlabConfig         `yaml:"gitlabConfig"`
}

type GithubConfig struct {
	APIToken string `yaml:"apiToken"`
}

type GitlabConfig struct {
	APIToken string `yaml:"token"`
	URL      string `yaml:"url"`
}

type LeaderElectionConfig struct {
	Enabled bool   `yaml:"enabled"`
	ID      string `yaml:"id"`
}

type ControllerTimers struct {
	DriftDetection     time.Duration `yaml:"driftDetection"`
	OnError            time.Duration `yaml:"onError"`
	WaitAction         time.Duration `yaml:"waitAction"`
	FailureGracePeriod time.Duration `yaml:"failureGracePeriod"`
}

type RepositoryConfig struct {
	SSHPrivateKey string `yaml:"sshPrivateKey"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
}

type RunnerConfig struct {
	Action                     string           `yaml:"action"`
	Layer                      Layer            `yaml:"layer"`
	Repository                 RepositoryConfig `yaml:"repository"`
	SSHKnownHostsConfigMapName string           `yaml:"sshKnowHostsConfigMapName"`
}

type TerragruntConfig struct {
	Enabled bool   `yaml:"enabled"`
	Version string `yaml:"version"`
}

type Layer struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type Redis struct {
	URL      string `yaml:"url"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

type Server struct {
	Addr    string        `yaml:"port"`
	Webhook WebhookConfig `yaml:"webhook"`
}

func (c *Config) Load(flags *pflag.FlagSet) error {
	v := viper.New()

	// burrito looks for configuration files called config.yaml, config.json,
	// config.toml, config.hcl, etc.
	v.SetConfigName("config")

	// burrito looks for configuration files in the common configuration
	// directories.
	v.AddConfigPath("/etc/burrito/")
	v.AddConfigPath("$HOME/.burrito/")

	// Viper logs the configuration file it uses, if any.
	if err := v.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", v.ConfigFileUsed())
	}

	// burrito can be configured with environment variables that start with
	// burrito_.
	v.SetEnvPrefix("burrito")
	v.AutomaticEnv()

	// Options with dashes in flag names have underscores when set inside a
	// configuration file or with environment variables.
	flags.SetNormalizeFunc(func(fs *pflag.FlagSet, name string) pflag.NormalizedName {
		name = strings.ReplaceAll(name, "-", "_")
		return pflag.NormalizedName(name)
	})
	v.BindPFlags(flags)

	// Nested configuration options set with environment variables use an
	// underscore as a separator.
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindEnvironmentVariables(v, *c)

	return v.Unmarshal(c)
}

// bindEnvironmentVariables inspects iface's structure and recursively binds its
// fields to environment variables. This is a workaround to a limitation of
// Viper, found here:
// https://github.com/spf13/viper/issues/188#issuecomment-399884438
func bindEnvironmentVariables(v *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		val := ifv.Field(i)
		typ := ift.Field(i)
		tv, ok := typ.Tag.Lookup("yaml")
		if !ok {
			continue
		}
		switch val.Kind() {
		case reflect.Struct:
			bindEnvironmentVariables(v, val.Interface(), append(parts, tv)...)
		default:
			v.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}

func TestConfig() *Config {
	return &Config{
		Redis: Redis{
			URL:      "redis://localhost:6379",
			Password: "",
			Database: 0,
		},
		Controller: ControllerConfig{
			Timers: ControllerTimers{
				DriftDetection:     20 * time.Minute,
				WaitAction:         5 * time.Minute,
				FailureGracePeriod: 15 * time.Second,
			},
		},
		Runner: RunnerConfig{
			SSHKnownHostsConfigMapName: "burrito-ssh-known-hosts",
		},
	}
}
