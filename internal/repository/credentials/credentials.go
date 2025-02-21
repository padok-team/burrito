package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/utils/url"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SharedCredentialsType = "credentials.burrito.tf/shared"
	CredentialsType       = "credentials.burrito.tf/repository"
)

type CredentialStore struct {
	TTL time.Duration
	mu  sync.Mutex
	client.Client
	sharedCredentials     []*SharedCredential
	repositoryCredentials []*RepositoryCredential
	lastUpdate            time.Time
}

func NewCredentialStore(client client.Client, ttl time.Duration) *CredentialStore {
	credentialStore := &CredentialStore{
		Client: client,
		TTL:    ttl,
	}
	err := credentialStore.updateCredentials()
	if err != nil {
		log.Errorf("Failed to update credentials: %v", err)
	}
	return credentialStore
}

func (s *CredentialStore) GetAllCredentials() ([]*SharedCredential, []*RepositoryCredential) {
	if time.Since(s.lastUpdate) < s.TTL {
		err := s.updateCredentials()
		if err != nil {
			log.Errorf("Failed to update credentials: %v", err)
		}
	}
	return s.sharedCredentials, s.repositoryCredentials
}

func (s *CredentialStore) updateCredentials() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Since(s.lastUpdate) < s.TTL {
		return nil
	}
	sharedSecrets := &corev1.SecretList{}
	err := s.List(context.Background(), sharedSecrets, client.MatchingFields{"type": SharedCredentialsType})
	if err != nil {
		s.lastUpdate = time.Now()
		return err
	}
	var sharedCredentials []*SharedCredential
	for _, secret := range sharedSecrets.Items {
		tmp, err := NewSharedCredentialsFromSecret(secret)
		if err != nil {
			log.New().Warnf("Failed to parse shared credentials from secret %s/%s", secret.Namespace, secret.Name)
			continue
		}
		sharedCredentials = append(sharedCredentials, tmp)
	}
	repositorySecrets := &corev1.SecretList{}
	err = s.List(context.Background(), repositorySecrets, client.MatchingFields{"type": CredentialsType})
	if err != nil {
		s.lastUpdate = time.Now()
		return err
	}
	var repositoryCredentials []*RepositoryCredential
	for _, secret := range repositorySecrets.Items {
		tmp, err := NewRepositoryCredentialsFromSecret(secret)
		if err != nil {
			log.New().Warnf("Failed to parse repository credentials from secret %s/%s", secret.Namespace, secret.Name)
			continue
		}
		repositoryCredentials = append(repositoryCredentials, tmp)
	}
	s.repositoryCredentials = repositoryCredentials
	s.sharedCredentials = sharedCredentials
	s.lastUpdate = time.Now()

	return nil
}

func (s *CredentialStore) GetCredentials(ctx context.Context, c client.Client, repository *configv1alpha1.TerraformRepository) (*Credential, error) {
	if time.Since(s.lastUpdate) < s.TTL {
		err := s.updateCredentials()
		if err != nil {
			log.Errorf("Failed to update credentials: %v", err)
		}
	}
	for _, repositoryCredentials := range s.repositoryCredentials {
		if repositoryCredentials.Matches(repository) {
			return &repositoryCredentials.Credentials, nil
		}
	}
	var sharedCredential *SharedCredential
	for _, tmp := range s.sharedCredentials {
		isAllowed := sharedCredential.IsAllowed(repository)
		matches := sharedCredential.Matches(repository)
		if isAllowed && matches {
			if sharedCredential != nil {
				if len(tmp.Credential.URL) > len(sharedCredential.Credential.URL) {
					sharedCredential = tmp
				}
			}
		}
	}
	if sharedCredential != nil {
		return &sharedCredential.Credential, nil
	}
	return nil, errors.New("no credentials found")
}

type SharedCredential struct {
	Credential     Credential
	AllowedTenants []string
}

type Credential struct {
	Provider string `json:"provider,omitempty"`
	// Basic auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// SSH auth
	SSHPrivateKey string `json:"sshPrivateKey,omitempty"`
	// GitHub App auth
	AppID             int64  `json:"appID,omitempty"`
	AppInstallationID int64  `json:"appInstallationID,omitempty"`
	AppPrivateKey     string `json:"appPrivateKey,omitempty"`
	// Token auth
	GitHubToken string `json:"githubToken,omitempty"`
	GitLabToken string `json:"gitlabToken,omitempty"`
	// Repository URL
	URL string `json:"url,omitempty"`
	// Mock provider
	EnableMock bool `json:"enableMock,omitempty"`
	// Secret for webhook handling
	WebhookSecret string `json:"webhookSecret,omitempty"`
}

type RepositoryCredential struct {
	Namespace   string
	Credentials Credential
}

func (c *RepositoryCredential) Matches(repository *configv1alpha1.TerraformRepository) bool {
	return url.NormalizeUrl(c.Credentials.URL) == url.NormalizeUrl(repository.Spec.Repository.Url) && c.Namespace == repository.Namespace
}

func NewRepositoryCredentialsFromSecret(secret corev1.Secret) (*RepositoryCredential, error) {
	credentials, err := parseRepositorySecret(secret)
	if err != nil {
		return nil, err
	}
	return &RepositoryCredential{
		Namespace:   secret.Namespace,
		Credentials: *credentials,
	}, nil
}

func NewSharedCredentialsFromSecret(secret corev1.Secret) (*SharedCredential, error) {
	credentials, err := parseRepositorySecret(secret)
	if err != nil {
		return nil, err
	}
	allowedTenants := strings.Split(secret.Annotations[annotations.AllowedTenants], ",")
	return &SharedCredential{
		Credential:     *credentials,
		AllowedTenants: allowedTenants,
	}, nil
}

func (t *SharedCredential) IsAllowed(repository *configv1alpha1.TerraformRepository) bool {
	if len(t.AllowedTenants) == 0 {
		return true
	}
	for _, allowedTenant := range t.AllowedTenants {
		if allowedTenant == repository.Namespace {
			return true
		}
	}
	return false
}

func (t *SharedCredential) Matches(repository *configv1alpha1.TerraformRepository) bool {
	return strings.Contains(url.NormalizeUrl(repository.Spec.Repository.Url), url.NormalizeUrl(t.Credential.URL))
}

func parseRepositorySecret(secret corev1.Secret) (*Credential, error) {
	raw, err := json.Marshal(secret.Data)
	if err != nil {
		return nil, err
	}
	repositorySecret := &Credential{}
	err = json.Unmarshal(raw, repositorySecret)
	if err != nil {
		return nil, err
	}
	return repositorySecret, nil
}
