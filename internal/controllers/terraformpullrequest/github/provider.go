package github

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v50/github"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Github struct {
	*github.Client
}

func (g *Github) IsConfigPresent(c *config.Config) bool {
	return c.Controller.GithubConfig.APIToken != ""
}

func (g *Github) Init(c *config.Config) error {
	ctx := context.Background()
	if !g.IsConfigPresent(c) {
		return errors.New("github config is not present")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Controller.GithubConfig.APIToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	g.Client = github.NewClient(tc)
	return nil
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
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	// Get all pull request files from Github
	var allChangedFiles []string
	for {
		changedFiles, resp, err := g.Client.PullRequests.ListFiles(context.TODO(), owner, repoName, id, nil)
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
	normalizedUrl := normalizeUrl(url)
	// nomalized url are "https://padok.github.com/owner/repo"
	// we remove "https://" then split on "/"
	split := strings.Split(normalizedUrl[8:], "/")
	return split[1], split[2]
}

func normalizeUrl(url string) string {
	if strings.Contains(url, "https://") {
		return url
	}
	// All SSH URL from GitHub are like "git@padok.github.com:<owner>/<repo>.git"
	// We split on ":" then remove ".git" by removing the last characters
	// To handle enterprise GitHub, we dynamically get "padok.github.com"
	// By removing "git@" at the beginning of the string
	split := strings.Split(url, ":")
	return fmt.Sprintf("https://%s/%s", split[0][4:], split[1][:len(split[1])-4])
}
