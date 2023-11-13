package gitlab

import (
	"fmt"
	"strconv"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	utils "github.com/padok-team/burrito/internal/utils/url"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type Gitlab struct {
	*gitlab.Client
}

func (g *Gitlab) IsConfigPresent(c *config.Config) bool {
	return c.Controller.GitlabConfig.APIToken != ""
}

func (g *Gitlab) Init(c *config.Config) error {
	if !g.IsConfigPresent(c) {
		return fmt.Errorf("gitlab config is not present")
	}
	client, err := gitlab.NewClient(c.Controller.GitlabConfig.APIToken, gitlab.WithBaseURL(c.Controller.GitlabConfig.URL))
	if err != nil {
		return err
	}
	g.Client = client
	return nil
}

func (g *Gitlab) IsFromProvider(pr *configv1alpha1.TerraformPullRequest) bool {
	return pr.Spec.Provider == "gitlab"
}

func (g *Gitlab) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	id, err := strconv.Atoi(pr.Spec.ID)
	if err != nil {
		log.Errorf("Error while parsing Gitlab merge request ID: %s", err)
		return []string{}, err
	}
	listOpts := gitlab.ListMergeRequestDiffsOptions{
		PerPage: 20,
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
		Body: gitlab.String(body),
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
