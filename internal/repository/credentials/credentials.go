package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
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

// credentialsSnapshot is an immutable view of the cached credentials. Each
// refresh builds a brand new snapshot and swaps it in atomically, so readers
// never mutate or block on it — they just load whatever snapshot is current.
type credentialsSnapshot struct {
	shared     []*SharedCredential
	repository []*RepositoryCredential
	updatedAt  time.Time
}

type CredentialStore struct {
	TTL time.Duration
	client.Client
	current atomic.Pointer[credentialsSnapshot]
	// refreshMu serializes refreshes so that concurrent stale readers trigger
	// a single API call instead of one each. It is never held by readers.
	refreshMu sync.Mutex
}

func NewCredentialStore(client client.Client, ttl time.Duration) *CredentialStore {
	credentialStore := &CredentialStore{
		Client: client,
		TTL:    ttl,
	}
	credentialStore.current.Store(&credentialsSnapshot{})
	return credentialStore
}

func (s *CredentialStore) GetAllCredentials() ([]*SharedCredential, []*RepositoryCredential) {
	snapshot := s.refreshIfNeeded()
	return snapshot.shared, snapshot.repository
}

// refreshIfNeeded returns the current snapshot, refreshing it first if it is
// older than the TTL. The common case (fresh snapshot) never takes a lock.
func (s *CredentialStore) refreshIfNeeded() *credentialsSnapshot {
	snapshot := s.load()
	if time.Since(snapshot.updatedAt) < s.TTL {
		return snapshot
	}

	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	// Re-check: another goroutine may have refreshed while we waited for the lock.
	snapshot = s.load()
	if time.Since(snapshot.updatedAt) < s.TTL {
		return snapshot
	}

	next, err := s.buildSnapshot(snapshot)
	if err != nil {
		log.Errorf("failed to update credentials: %v", err)
	}
	// Stamp the snapshot even when the refresh failed, so a failing API call is
	// retried after the TTL instead of on every single request.
	next.updatedAt = time.Now()
	s.current.Store(next)
	return next
}

// load returns the current snapshot, never nil. A CredentialStore built as a
// struct literal instead of through NewCredentialStore has no snapshot yet.
func (s *CredentialStore) load() *credentialsSnapshot {
	if snapshot := s.current.Load(); snapshot != nil {
		return snapshot
	}
	return &credentialsSnapshot{}
}

// buildSnapshot lists secrets and returns the next snapshot. On a list error it
// returns a snapshot carrying the previous credentials, so a failed refresh
// keeps serving the last known-good values instead of dropping them.
func (s *CredentialStore) buildSnapshot(previous *credentialsSnapshot) (*credentialsSnapshot, error) {
	next := &credentialsSnapshot{
		shared:     previous.shared,
		repository: previous.repository,
	}

	sharedSecrets := &corev1.SecretList{}
	err := s.List(context.Background(), sharedSecrets, client.MatchingFields{"type": SharedCredentialsType})
	if err != nil {
		return next, err
	}
	var sharedCredentials []*SharedCredential
	for _, secret := range sharedSecrets.Items {
		tmp, err := NewSharedCredentialsFromSecret(secret)
		if err != nil {
			log.Errorf("failed to parse shared credentials from secret %s/%s: %s", secret.Namespace, secret.Name, err)
			continue
		}
		sharedCredentials = append(sharedCredentials, tmp)
	}

	repositorySecrets := &corev1.SecretList{}
	err = s.List(context.Background(), repositorySecrets, client.MatchingFields{"type": CredentialsType})
	if err != nil {
		return next, err
	}
	var repositoryCredentials []*RepositoryCredential
	for _, secret := range repositorySecrets.Items {
		tmp, err := NewRepositoryCredentialsFromSecret(secret)
		if err != nil {
			log.Errorf("failed to parse repository credentials from secret %s/%s: %s", secret.Namespace, secret.Name, err)
			continue
		}
		repositoryCredentials = append(repositoryCredentials, tmp)
	}

	next.shared = sharedCredentials
	next.repository = repositoryCredentials
	return next, nil
}

// Returns the credentials for a given repository. If a specific repository credential is found, it will be returned.
// If not, the most specific shared credential that matches the repository will be returned.
func (s *CredentialStore) GetCredentials(repository *configv1alpha1.TerraformRepository) (*Credential, error) {
	snapshot := s.refreshIfNeeded()
	for _, repositoryCredentials := range snapshot.repository {
		if repositoryCredentials.Matches(repository) {
			return &repositoryCredentials.Credential, nil
		}
	}
	var sharedCredential *SharedCredential
	for _, tmp := range snapshot.shared {
		isAllowed := tmp.IsAllowed(repository)
		matches := tmp.Matches(repository)
		if isAllowed && matches {
			if sharedCredential != nil {
				// Check if the current shared credential (`tmp`) is more specific than the previous one (`sharedCredential`)
				if len(sharedCredential.Credential.URL) < len(tmp.Credential.URL) {
					sharedCredential = tmp
				}
			} else {
				sharedCredential = tmp
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
	GitHubAppID             string `json:"githubAppID,omitempty"`
	GitHubAppInstallationID string `json:"githubAppInstallationID,omitempty"`
	GitHubAppPrivateKey     string `json:"githubAppPrivateKey,omitempty"`
	// Token auth
	GitHubToken string `json:"githubToken,omitempty"`
	GitLabToken string `json:"gitlabToken,omitempty"`
	// Repository URL
	URL string `json:"url,omitempty"`
	// Secret for webhook handling
	WebhookSecret string `json:"webhookSecret,omitempty"`
}

type RepositoryCredential struct {
	Namespace  string
	Credential Credential
}

func (c *RepositoryCredential) Matches(repository *configv1alpha1.TerraformRepository) bool {
	return url.NormalizeUrl(c.Credential.URL) == url.NormalizeUrl(repository.Spec.Repository.Url) && c.Namespace == repository.Namespace
}

func NewRepositoryCredentialsFromSecret(secret corev1.Secret) (*RepositoryCredential, error) {
	credential, err := parseRepositorySecret(secret)
	if err != nil {
		return nil, err
	}
	return &RepositoryCredential{
		Namespace:  secret.Namespace,
		Credential: *credential,
	}, nil
}

func NewSharedCredentialsFromSecret(secret corev1.Secret) (*SharedCredential, error) {
	credential, err := parseRepositorySecret(secret)
	if err != nil {
		return nil, err
	}
	allowedTenants := []string{}
	value, ok := secret.Annotations[annotations.AllowedTenants]
	if ok {
		allowedTenants = strings.Split(value, ",")
	}
	return &SharedCredential{
		Credential:     *credential,
		AllowedTenants: allowedTenants,
	}, nil
}

func (t *SharedCredential) IsAllowed(repository *configv1alpha1.TerraformRepository) bool {
	if len(t.AllowedTenants) == 0 {
		return true
	}
	return slices.Contains(t.AllowedTenants, repository.Namespace)
}

func (t *SharedCredential) Matches(repository *configv1alpha1.TerraformRepository) bool {
	return strings.Contains(url.NormalizeUrl(repository.Spec.Repository.Url), url.NormalizeUrl(t.Credential.URL))
}

func parseRepositorySecret(secret corev1.Secret) (*Credential, error) {
	unencoded := make(map[string]string)
	for key, value := range secret.Data {
		unencoded[key] = string(value)
	}

	raw, err := json.Marshal(unencoded)
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
