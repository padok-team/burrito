package runner

import (
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/padok-team/burrito/internal/burrito/config"
	log "github.com/sirupsen/logrus"
)

func clone(repository config.RepositoryConfig, branch, path string) (*git.Repository, error) {
	cloneOptions, err := getCloneOptions(repository, branch, path)
	if err != nil {
		return &git.Repository{}, err
	}
	return git.PlainClone(WorkingDir, false, cloneOptions)
}

func getCloneOptions(repository config.RepositoryConfig, branch, path string) (*git.CloneOptions, error) {
	authMethod := "ssh"
	cloneOptions := &git.CloneOptions{
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		URL:           repository.URL,
	}
	if strings.Contains(repository.URL, "https://") {
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
