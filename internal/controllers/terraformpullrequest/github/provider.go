package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v66/github"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	utils "github.com/padok-team/burrito/internal/utils/url"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Github struct {
	*github.Client
	AppId             string
	AppInstallationId string
	AppPrivateKey     string
	ApiToken          string
	Url               string
}

type GitHubSubscription string

const (
	GitHubEnterprise GitHubSubscription = "enterprise"
	GitHubClassic    GitHubSubscription = "classic"
)

func (g *Github) IsAppConfigPresent() bool {
	return g.AppId != "" && g.AppInstallationId != "" && g.AppPrivateKey != ""
}

func (g *Github) IsAPITokenConfigPresent() bool {
	return g.ApiToken != ""
}

func (g *Github) Init() error {
	apiUrl, subscription, err := inferBaseURL(g.Url)
	httpClient := &http.Client{}
	if g.IsAppConfigPresent() {
		appId, err := strconv.ParseInt(g.AppId, 10, 64)
		if err != nil {
			return errors.New("error while parsing github app id: " + err.Error())
		}
		appInstallationId, err := strconv.ParseInt(g.AppInstallationId, 10, 64)
		if err != nil {
			return errors.New("error while parsing github app installation id: " + err.Error())
		}
		itr, err := ghinstallation.New(http.DefaultTransport, appId, appInstallationId, []byte(g.AppPrivateKey))
		if err != nil {
			return errors.New("error while creating github installation client: " + err.Error())
		}
		httpClient.Transport = itr
	} else if g.IsAPITokenConfigPresent() {
		ctx := context.Background()

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: g.ApiToken},
		)
		httpClient = oauth2.NewClient(ctx, ts)
	} else {
		return errors.New("github config is not present")
	}

	if subscription == GitHubEnterprise {
		g.Client, err = github.NewClient(httpClient).WithEnterpriseURLs(apiUrl, apiUrl)
		if err != nil {
			return errors.New("error while creating github enterprise client: " + err.Error())
		}
	} else if subscription == GitHubClassic {
		g.Client = github.NewClient(httpClient)
	}
	return nil
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

func inferBaseURL(repoURL string) (string, GitHubSubscription, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid repository URL: %w", err)
	}

	host := parsedURL.Host
	host = strings.TrimPrefix(host, "www.")

	if host != "github.com" {
		return fmt.Sprintf("https://%s/api/v3", host), GitHubEnterprise, nil
	} else {
		return "", GitHubClassic, nil
	}
}
