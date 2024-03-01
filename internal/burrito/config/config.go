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
	Runner     RunnerConfig     `mapstructure:"runner"`
	Controller ControllerConfig `mapstructure:"controller"`
	Redis      Redis            `mapstructure:"redis"`
	Server     ServerConfig     `mapstructure:"server"`
}

type WebhookConfig struct {
	Github WebhookGithubConfig `mapstructure:"github"`
	Gitlab WebhookGitlabConfig `mapstructure:"gitlab"`
}

type WebhookGithubConfig struct {
	Secret string `mapstructure:"secret"`
}

type WebhookGitlabConfig struct {
	Secret string `mapstructure:"secret"`
}

type ControllerConfig struct {
	Namespaces             []string             `mapstructure:"namespaces"`
	Timers                 ControllerTimers     `mapstructure:"timers"`
	TerraformMaxRetries    int                  `mapstructure:"terraformMaxRetries"`
	Types                  []string             `mapstructure:"types"`
	LeaderElection         LeaderElectionConfig `mapstructure:"leaderElection"`
	MetricsBindAddress     string               `mapstructure:"metricsBindAddress"`
	HealthProbeBindAddress string               `mapstructure:"healthProbeBindAddress"`
	KubernetesWebhookPort  int                  `mapstructure:"kubernetesWebhookPort"`
	GithubConfig           GithubConfig         `mapstructure:"githubConfig"`
	GitlabConfig           GitlabConfig         `mapstructure:"gitlabConfig"`
}

type GithubConfig struct {
	AppId          int64  `mapstructure:"appId"`
	InstallationId int64  `mapstructure:"installationId"`
	PrivateKey     string `mapstructure:"privateKey"`
	APIToken       string `mapstructure:"apiToken"`
}

type GitlabConfig struct {
	APIToken string `mapstructure:"apiToken"`
	URL      string `mapstructure:"url"`
}

type LeaderElectionConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	ID      string `mapstructure:"id"`
}

type ControllerTimers struct {
	DriftDetection     time.Duration `mapstructure:"driftDetection"`
	OnError            time.Duration `mapstructure:"onError"`
	WaitAction         time.Duration `mapstructure:"waitAction"`
	FailureGracePeriod time.Duration `mapstructure:"failureGracePeriod"`
}

type RepositoryConfig struct {
	SSHPrivateKey string `mapstructure:"sshPrivateKey"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
}

type RunnerConfig struct {
	Action                     string           `mapstructure:"action"`
	Layer                      Layer            `mapstructure:"layer"`
	Repository                 RepositoryConfig `mapstructure:"repository"`
	SSHKnownHostsConfigMapName string           `mapstructure:"sshKnownHostsConfigMapName"`
	RunnerBinaryPath           string           `mapstructure:"runnerBinaryPath"`
}

type Layer struct {
	Name      string `mapstructure:"name"`
	Namespace string `mapstructure:"namespace"`
}

type Redis struct {
	Hostname   string `mapstructure:"hostname"`
	ServerPort int    `mapstructure:"serverPort"`
	Password   string `mapstructure:"password"`
	Database   int    `mapstructure:"database"`
}

type ServerConfig struct {
	Addr    string        `mapstructure:"addr"`
	Webhook WebhookConfig `mapstructure:"webhook"`
}

func (c *Config) Load(flags *pflag.FlagSet) error {
	v := viper.New()

	// burrito looks for configuration files called config.yaml, config.json,
	// config.toml, config.hcl, etc.
	v.SetConfigName("config")

	// burrito looks for configuration files in the common configuration
	// directories, as well as in the current directory.
	v.AddConfigPath("/etc/burrito/")
	v.AddConfigPath("$HOME/.burrito/")
	v.AddConfigPath(".")

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
	err := v.BindPFlags(flags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error binding flags: %s\n", err)
		return err
	}

	// Nested configuration options set with environment variables use an
	// underscore as a separator.
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err = bindEnvironmentVariables(v, *c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error binding environment variables: %s\n", err)
		return err
	}

	return v.Unmarshal(c)
}

// bindEnvironmentVariables inspects iface's structure and recursively binds its
// fields to environment variables. This is a workaround to a limitation of
// Viper, found here:
// https://github.com/spf13/viper/issues/188#issuecomment-399884438
func bindEnvironmentVariables(v *viper.Viper, iface interface{}, parts ...string) error {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		val := ifv.Field(i)
		typ := ift.Field(i)
		tv, ok := typ.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch val.Kind() {
		case reflect.Struct:
			err := bindEnvironmentVariables(v, val.Interface(), append(parts, tv)...)
			if err != nil {
				return err
			}
		default:
			err := v.BindEnv(strings.Join(append(parts, tv), "."))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func TestConfig() *Config {
	return &Config{
		Redis: Redis{
			Hostname:   "localhost",
			ServerPort: 6379,
			Password:   "",
			Database:   0,
		},
		Controller: ControllerConfig{
			TerraformMaxRetries: 5,
			Timers: ControllerTimers{
				DriftDetection:     20 * time.Minute,
				WaitAction:         5 * time.Minute,
				FailureGracePeriod: 15 * time.Second,
				OnError:            1 * time.Minute,
			},
		},
		Runner: RunnerConfig{
			SSHKnownHostsConfigMapName: "burrito-ssh-known-hosts",
		},
	}
}
