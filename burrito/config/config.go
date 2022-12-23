package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Runner     RunnerConfig     `yaml:"runner"`
	Controller ControllerConfig `yaml:"controller"`
	Redis      Redis            `yaml:"redis"`
}

type ControllerConfig struct {
	WatchedNamespaces []string         `yaml:"namespaces"`
	Timers            ControllerTimers `yaml:"timers"`
}

type ControllerTimers struct {
	DriftDetection int `yaml:"driftDetection"`
	WaitAction     int `yaml:"waitAction"`
	OnError        int `yaml:"onError"`
}

type RepositoryConfig struct {
	URL      string `yaml:"url"`
	SSH      string `yaml:"ssh"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type RunnerConfig struct {
	Path       string           `yaml:"path"`
	Branch     string           `yaml:"branch"`
	Version    string           `yaml:"version"`
	Action     string           `yaml:"action"`
	Repository RepositoryConfig `yaml:"repository"`
	Layer      LayerConfig      `yaml:"layer"`
}

type LayerConfig struct {
	Lock     string `yaml:"lock"`
	PlanSum  string `yaml:"planSum"`
	PlanBin  string `yaml:"planBin"`
	ApplySum string `yaml:"applySum"`
	PlanDate string `yaml:"planDate"`
}

type Redis struct {
	URL      string `yaml:"url"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
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
