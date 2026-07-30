package bitbucket

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/repository/credentials"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BitbucketClient struct {
	username string
	token    string
	password string
	baseURL  string
}

func (c *BitbucketClient) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.SetBasicAuth(c.username, c.token)
	} else {
		req.SetBasicAuth(c.username, c.password)
	}
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

type APIProvider struct {
	config credentials.Credential
	client *BitbucketClient
}

func (api *APIProvider) GetChanges(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ([]string, error) {
	workspace, repoSlug := getBitbucketNamespacedName(repository.Spec.Repository.Url)
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%s/diffstat",
		workspace, repoSlug, pr.Spec.ID)
	resp, err := api.client.doRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error getting Bitbucket PR changes: %s", err)
		return []string{}, err
	}
	defer resp.Body.Close()
	var result struct {
		Values []struct {
			New struct {
				Path string `json:"path"`
			} `json:"new"`
		} `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []string{}, err
	}
	var changes []string
	for _, v := range result.Values {
		changes = append(changes, v.New.Path)
	}
	return changes, nil
}

func (api *APIProvider) Comment(repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, prComment comment.Comment) error {
	body, err := prComment.Generate("")
	if err != nil {
		return err
	}
	body = comment.WithManagedMarker(body)
	workspace, repoSlug := getBitbucketNamespacedName(repository.Spec.Repository.Url)
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests/%s/comments",
		workspace, repoSlug, pr.Spec.ID)
	payload := map[string]interface{}{"content": map[string]interface{}{"raw": body}}
	payloadBytes, _ := json.Marshal(payload)
	resp, err := api.client.doRequest("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("Bitbucket comment failed: %d", resp.StatusCode)
	}
	return nil
}

func (api *APIProvider) ListPullRequests(repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformPullRequest, error) {
	workspace, repoSlug := getBitbucketNamespacedName(repository.Spec.Repository.Url)
	url := fmt.Sprintf("https://api.bitbucket.org/2.0/repositories/%s/%s/pullrequests?state=OPEN",
		workspace, repoSlug)
	resp, err := api.client.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Values []struct {
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Source struct {
				Branch struct {
					Name string `json:"name"`
				} `json:"branch"`
				Commit struct {
					Hash string `json:"hash"`
				} `json:"commit"`
			} `json:"source"`
			Destination struct {
				Branch struct {
					Name string `json:"name"`
				} `json:"branch"`
			} `json:"destination"`
		} `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	var pullRequests []configv1alpha1.TerraformPullRequest
	for _, pr := range result.Values {
		pullRequests = append(pullRequests, configv1alpha1.TerraformPullRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", repository.Name, pr.ID),
				Namespace: repository.Namespace,
			},
			Spec: configv1alpha1.TerraformPullRequestSpec{
				Branch: pr.Source.Branch.Name,
				Base:   pr.Destination.Branch.Name,
				ID:     fmt.Sprintf("%d", pr.ID),
				Repository: configv1alpha1.TerraformLayerRepository{
					Name:      repository.Name,
					Namespace: repository.Namespace,
				},
			},
		})
	}
	return pullRequests, nil
}

func getBitbucketNamespacedName(url string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "bitbucket.org/"), "/")
	if len(parts) >= 2 {
		return parts[0], strings.TrimSuffix(parts[1], ".git")
	}
	return "", ""
}
