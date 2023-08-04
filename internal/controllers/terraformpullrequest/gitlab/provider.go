package gitlab

import (
	"fmt"
	"strconv"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
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
	getOpts := gitlab.GetMergeRequestChangesOptions{
		AccessRawDiffs: gitlab.Bool(true),
	}

	mr, _, err := g.Client.MergeRequests.GetMergeRequestChanges(getGitlabNamespacedName(repository.Spec.Repository.Url), id, &getOpts)
	if err != nil {
		log.Errorf("Error while getting merge request changes: %s", err)
		return []string{}, err
	}
	var changes []string
	for _, change := range mr.Changes {
		changes = append(changes, change.NewPath)
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
	normalizedUrl := normalizeUrl(url)
	return strings.Join(strings.Split(normalizedUrl[8:], "/")[1:], "/")
}

func normalizeUrl(url string) string {
	if strings.Contains(url, "https://") {
		return url
	}
	// All SSH URL from GitLab are like "git@<URL>:<owner>/<repo>.git"
	// We split on ":" then remove ".git" by removing the last characters
	// To handle enterprise GitLab on premise, we dynamically get "padok.gitlab.com"
	// By removing "git@" at the beginning of the string
	split := strings.Split(url, ":")
	return fmt.Sprintf("https://%s/%s", split[0][4:], split[1][:len(split[1])-4])
}
