package runner

import (
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	log "github.com/sirupsen/logrus"
)

// Fetch the content of the specified repository on the specified branch with git clone
//
// TODO: Fetch repo from datastore when repository controller is implemented
func FetchRepositoryContent(repo *configv1alpha1.TerraformRepository, branch string, config config.RepositoryConfig) (*git.Repository, error) {
	log.Infof("fetching repository %s on %s branch with git clone", repo.Spec.Repository.Url, branch)
	cloneOptions, err := getCloneOptions(config, repo.Spec.Repository.Url, branch)
	if err != nil {
		return &git.Repository{}, err
	}
	return git.PlainClone(RepositoryDir, false, cloneOptions)
}

func getCloneOptions(repository config.RepositoryConfig, URL, branch string) (*git.CloneOptions, error) {
	authMethod := "ssh"
	cloneOptions := &git.CloneOptions{
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		URL:           URL,
	}
	if strings.Contains(URL, "https://") {
		authMethod = "https"
	}
	log.Infof("clone method is %s", authMethod)
	switch authMethod {
	case "ssh":
		if repository.SSHPrivateKey == "" {
			log.Infof("detected keyless authentication")
			return cloneOptions, nil
		}
		log.Infof("private key found")
		publicKeys, err := ssh.NewPublicKeys("git", []byte(repository.SSHPrivateKey), "")
		if err != nil {
			return cloneOptions, err
		}
		cloneOptions.Auth = publicKeys

	case "https":
		if repository.Username != "" && repository.Password != "" {
			log.Infof("username and password found")
			cloneOptions.Auth = &http.BasicAuth{
				Username: repository.Username,
				Password: repository.Password,
			}
		} else {
			log.Infof("passwordless authentication detected")
		}
	}
	return cloneOptions, nil
}
