package gitlab

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	utils "github.com/padok-team/burrito/internal/utils/url"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type Gitlab struct {
	*gitlab.Client
	ApiToken string
	Url      string
}

func (g *Gitlab) Init() error {
	apiUrl, err := inferBaseURL(utils.NormalizeUrl(g.Url))
	if err != nil {
		return err
	}
	client, err := gitlab.NewClient(g.ApiToken, gitlab.WithBaseURL(apiUrl))
	if err != nil {
		return err
	}
	g.Client = client
	return nil
}

func (g *Gitlab) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Gitlab merge request ID: %s", err)
		return []string{}, err
	}
	listOpts := gitlab.ListMergeRequestDiffsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
		},
	}
	var changes []string
	for {
		diffs, resp, err := g.Client.MergeRequests.ListMergeRequestDiffs(getGitlabNamespacedName(repository.Spec.Repository.Url), id, &listOpts)
		if err != nil {
			log.Errorf("Error while getting merge request changes: %s", err)
			return []string{}, err
		}
		for _, change := range diffs {
			changes = append(changes, change.NewPath)
		}
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}
	return changes, nil
}

func (g *Gitlab) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, comment comment.Comment) error {
	body, err := comment.Generate(pr.Annotations[annotations.LastBranchCommit])
	if err != nil {
		log.Errorf("Error while generating comment: %s", err)
		return err
	}
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Gitlab merge request ID: %s", err)
		return err
	}
	_, _, err = g.Client.Notes.CreateMergeRequestNote(getGitlabNamespacedName(repository.Spec.Repository.Url), id, &gitlab.CreateMergeRequestNoteOptions{
		Body: gitlab.Ptr(body),
	})
	if err != nil {
		log.Errorf("Error while creating merge request note: %s", err)
		return err
	}
	return nil
}

func getGitlabNamespacedName(url string) string {
	normalizedUrl := utils.NormalizeUrl(url)
	return strings.Join(strings.Split(normalizedUrl[8:], "/")[1:], "/")
}

func inferBaseURL(repoURL string) (string, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("invalid repository URL: %w", err)
	}

	host := parsedURL.Host
	host = strings.TrimPrefix(host, "www.")
	return fmt.Sprintf("https://%s/api/v4", host), nil
}
