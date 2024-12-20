package gitprovider

import (
	"fmt"
	"slices"

	"github.com/padok-team/burrito/internal/utils/gitprovider/github"
	"github.com/padok-team/burrito/internal/utils/gitprovider/gitlab"
	"github.com/padok-team/burrito/internal/utils/gitprovider/mock"
	"github.com/padok-team/burrito/internal/utils/gitprovider/standard"
	"github.com/padok-team/burrito/internal/utils/gitprovider/types"
	log "github.com/sirupsen/logrus"
)

type Provider = types.Provider
type Config = types.Config

var providers = map[string]struct {
	IsAvailable func(types.Config, []string) bool
	create      func(types.Config) types.Provider
	priority    int64
}{
	"github":   {github.IsAvailable, func(config types.Config) types.Provider { return &github.Github{Config: config} }, 0},
	"gitlab":   {gitlab.IsAvailable, func(config types.Config) types.Provider { return &gitlab.Gitlab{Config: config} }, 1},
	"mock":     {mock.IsAvailable, func(types.Config) types.Provider { return &mock.Mock{} }, 99},
	"standard": {standard.IsAvailable, func(config types.Config) types.Provider { return &standard.Standard{Config: config} }, 100},
}

// New creates a new git provider based on the given configuration and capabilities.
// The provider is selected based on the given capabilities and the provider's priority.
func New(config types.Config, capabilities []string) (types.Provider, error) {
	available, err := ListAvailable(config, capabilities)
	if err != nil || len(available) == 0 {
		return nil, fmt.Errorf("No git provider available with the given configuration")
	}
	log.Infof("Creating git provider of type %s", available[0])
	return NewWithName(config, available[0])
}

// NewWithName creates a new git provider based on the given configuration and provider name.
// Caution: this function does not check if the provider is available with the given configuration or capabilities.
// It is the caller's responsibility to ensure that the provider is available. Use ListAvailable for that purpose.
func NewWithName(config types.Config, providerName string) (types.Provider, error) {
	if provider, exists := providers[providerName]; exists {
		return provider.create(config), nil
	}
	return nil, fmt.Errorf("unknown provider %s", providerName)
}

// ListAvailable returns a list of providers that:
// - match the requested capabilities
// - can be initialized with the given configuration
func ListAvailable(config types.Config, capabilities []string) ([]string, error) {
	var availableProviders []string
	for name, provider := range providers {
		if provider.IsAvailable(config, capabilities) {
			availableProviders = append(availableProviders, name)
		}
	}

	slices.SortFunc(availableProviders, func(a, b string) int {
		if providers[a].priority < providers[b].priority {
			return -1
		} else if providers[a].priority > providers[b].priority {
			return 1
		}
		return 0
	})

	return availableProviders, nil
}
