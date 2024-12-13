package runner

import (
	"context"
	nethttp "net/http"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	log "github.com/sirupsen/logrus"
)

type CloneOptions struct {
	AuthenticationType string
	CloneOptions       *git.CloneOptions
}

// Fetch the content of the specified repository on the specified branch with git clone
//
// TODO: Fetch repo from datastore when repository controller is implemented
func FetchRepositoryContent(repo *configv1alpha1.TerraformRepository, branch string, config *config.Config) (*git.Repository, error) {
	log.Infof("fetching repository %s on %s branch with git clone", repo.Spec.Repository.Url, branch)
	cloneOptionsList, err := getCloneOptionsList(config.Runner.Repository, repo.Spec.Repository.Url, branch)
	if err != nil {
		return &git.Repository{}, err
	}

	var lastErr error
	for _, cloneOptions := range cloneOptionsList {
		options := cloneOptions.CloneOptions
		log.Infof("trying to clone with %s authentication", cloneOptions.AuthenticationType)
		repo, err := git.PlainClone(config.Runner.RepositoryPath, false, options)
		if err == nil {
			return repo, nil
		}
		lastErr = err
		log.Warnf("clone attempt failed: %v", err)
	}

	return &git.Repository{}, lastErr
}

func getCloneOptionsList(repository config.RepositoryConfig, URL, branch string) ([]*CloneOptions, error) {
	cloneOptions := &git.CloneOptions{
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		URL:           URL,
	}

	var cloneOptionsList []*CloneOptions

	if strings.Contains(URL, "https://") {
		// HTTPS cloning methods
		if repository.Username != "" && repository.Password != "" {
			log.Infof("Git username and password found in repository secret")
			cloneOptionsList = append(cloneOptionsList, &CloneOptions{
				AuthenticationType: "username-password",
				CloneOptions: &git.CloneOptions{
					ReferenceName: cloneOptions.ReferenceName,
					URL:           cloneOptions.URL,
					Auth: &http.BasicAuth{
						Username: repository.Username,
						Password: repository.Password,
					},
				},
			})
		}
		if repository.GithubAppId != 0 && repository.GithubAppInstallationId != 0 && repository.GithubAppPrivateKey != "" {
			log.Infof("Github app credentials found in repository secret")
			tr, err := ghinstallation.New(
				nethttp.DefaultTransport,
				repository.GithubAppId,
				repository.GithubAppInstallationId,
				[]byte(repository.GithubAppPrivateKey),
			)
			if err == nil {
				token, err := tr.Token(context.Background())
				if err == nil {
					cloneOptionsList = append(cloneOptionsList, &CloneOptions{
						AuthenticationType: "github-app",
						CloneOptions: &git.CloneOptions{
							ReferenceName: cloneOptions.ReferenceName,
							URL:           cloneOptions.URL,
							Auth: &http.BasicAuth{
								Username: "x-access-token",
								Password: token,
							},
						},
					})
				} else {
					log.Warnf("failed to create Github app token from credentials: %v", err)
				}
			}
		}
		if repository.GithubToken != "" {
			log.Infof("Github token found in repository secret")
			cloneOptionsList = append(cloneOptionsList, &CloneOptions{
				AuthenticationType: "github-token",
				CloneOptions: &git.CloneOptions{
					ReferenceName: cloneOptions.ReferenceName,
					URL:           cloneOptions.URL,
					Auth: &http.BasicAuth{
						Username: "x-access-token",
						Password: repository.GithubToken,
					},
				},
			})
		}
		if repository.GitlabToken != "" {
			log.Infof("Gitlab token found in repository secret")
			cloneOptionsList = append(cloneOptionsList, &CloneOptions{
				AuthenticationType: "gitlab-token",
				CloneOptions: &git.CloneOptions{
					ReferenceName: cloneOptions.ReferenceName,
					URL:           cloneOptions.URL,
					Auth: &http.BasicAuth{
						Password: repository.GitlabToken,
					},
				},
			})
		}
		cloneOptionsList = append(cloneOptionsList, &CloneOptions{
			AuthenticationType: "passwordless",
			CloneOptions:       cloneOptions,
		})
	} else {
		// SSH cloning method
		if repository.SSHPrivateKey != "" {
			log.Infof("adding SSH private key authentication")
			publicKeys, err := ssh.NewPublicKeys("git", []byte(repository.SSHPrivateKey), "")
			if err == nil {
				cloneOptions.Auth = publicKeys
			}
		}
		cloneOptionsList = append(cloneOptionsList, &CloneOptions{
			AuthenticationType: "ssh",
			CloneOptions:       cloneOptions,
		})
	}

	return cloneOptionsList, nil
}
