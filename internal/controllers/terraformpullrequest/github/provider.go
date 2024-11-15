package github

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v66/github"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	utils "github.com/padok-team/burrito/internal/utils/url"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Github struct {
	*github.Client
}

func (g *Github) IsAppConfigPresent(c *config.Config) bool {
	return c.Controller.GithubConfig.AppId != 0 && c.Controller.GithubConfig.InstallationId != 0 && len(c.Controller.GithubConfig.PrivateKey) != 0
}

func (g *Github) IsAPITokenConfigPresent(c *config.Config) bool {
	return len(c.Controller.GithubConfig.APIToken) != 0
}

func (g *Github) Init(c *config.Config) error {
	if g.IsAppConfigPresent(c) {
		itr, err := ghinstallation.New(http.DefaultTransport, c.Controller.GithubConfig.AppId, c.Controller.GithubConfig.InstallationId, []byte(c.Controller.GithubConfig.PrivateKey))

		if err != nil {
			return errors.New("error while creating github installation client: " + err.Error())
		}

		g.Client = github.NewClient(&http.Client{Transport: itr})
		return nil
	} else if g.IsAPITokenConfigPresent(c) {
		ctx := context.Background()

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: c.Controller.GithubConfig.APIToken},
		)
		tc := oauth2.NewClient(ctx, ts)

		g.Client = github.NewClient(tc)
		return nil
	} else {
		return errors.New("github config is not present")
	}
}

func (g *Github) IsFromProvider(pr *configv1alpha1.TerraformPullRequest) bool {
	return pr.Spec.Provider == "github"
}

func (g *Github) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	owner, repoName := parseGithubUrl(repository.Spec.Repository.Url)
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Github pull request ID: %s", err)
		return []string{}, err
	}
	// Per page is 30 by default, max is 100
	opts := &github.ListOptions{
		PerPage: 100,
	}
	// Get all pull request files from Github
	var allChangedFiles []string
	for {
		changedFiles, resp, err := g.Client.PullRequests.ListFiles(context.TODO(), owner, repoName, id, opts)
		if err != nil {
			return []string{}, err
		}
		for _, file := range changedFiles {
			if *file.Status != "unchanged" {
				allChangedFiles = append(allChangedFiles, *file.Filename)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allChangedFiles, nil
}

func (g *Github) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, comment comment.Comment) error {
	body, err := comment.Generate(pr.Annotations[annotations.LastBranchCommit])
	if err != nil {
		log.Errorf("Error while generating comment: %s", err)
		return err
	}
	owner, repoName := parseGithubUrl(repository.Spec.Repository.Url)
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Github pull request ID: %s", err)
		return err
	}
	_, _, err = g.Client.Issues.CreateComment(context.TODO(), owner, repoName, id, &github.IssueComment{
		Body: &body,
	})
	return err
}

func parseGithubUrl(url string) (string, string) {
	normalizedUrl := utils.NormalizeUrl(url)
	// nomalized url are "https://padok.github.com/owner/repo"
	// we remove "https://" then split on "/"
	split := strings.Split(normalizedUrl[8:], "/")
	return split[1], split[2]
}
