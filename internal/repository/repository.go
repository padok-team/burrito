package repository

import (
	"fmt"

	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/repository/providers/github"
	"github.com/padok-team/burrito/internal/repository/providers/gitlab"
	"github.com/padok-team/burrito/internal/repository/types"
)

func GetProvider(RepositoryCredentials credentials.Credential) (types.Provider, error) {
	switch RepositoryCredentials.Provider {
	case "github":
		return &github.Github{Config: RepositoryCredentials}, nil
	case "gitlab":
		return &gitlab.Gitlab{Config: RepositoryCredentials}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", RepositoryCredentials.Provider)
	}
}
