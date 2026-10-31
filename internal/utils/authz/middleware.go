package authz

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/patrickmn/go-cache"
	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type (
	Authz struct {
		Cache           *cache.Cache
		Client          client.Clientset
		Audience        string
		ServiceAccounts []string
		// TenantServiceAccounts maps a tenant service account username to the only namespace it may access
		TenantServiceAccounts map[string]string
	}
)

func NewAuthz() *Authz {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := client.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &Authz{
		Cache:  cache.New(5*time.Minute, 10*time.Minute),
		Client: *clientset,
	}
}

func (a *Authz) SetAudience(audience string) {
	a.Audience = audience
}

func (a *Authz) AddServiceAccount(namespace string, name string) {
	a.ServiceAccounts = append(a.ServiceAccounts, fmt.Sprintf("system:serviceaccount:%s:%s", namespace, name))
}

// AddTenantServiceAccount allows a service account restricted to requests on its own namespace.
func (a *Authz) AddTenantServiceAccount(namespace string, name string) {
	if a.TenantServiceAccounts == nil {
		a.TenantServiceAccounts = map[string]string{}
	}
	a.TenantServiceAccounts[fmt.Sprintf("system:serviceaccount:%s:%s", namespace, name)] = namespace
}

// Process is the middleware function.
func (a *Authz) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		authz := c.Request().Header.Get("Authorization")
		if authz == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing Authorization header")
		}
		cached, found := a.Cache.Get(authz)
		username, _ := cached.(string)
		if !found {
			tr := &v1.TokenReview{
				Spec: v1.TokenReviewSpec{
					Token:     authz,
					Audiences: []string{a.Audience},
				},
			}
			result, err := a.Client.AuthenticationV1().TokenReviews().Create(context.Background(), tr, metav1.CreateOptions{})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "error validating token")
			}
			if !result.Status.Authenticated {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			username = result.Status.User.Username
			if !a.isServiceAccountAllowed(username) {
				return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized user")
			}
			a.Cache.Set(authz, username, cache.DefaultExpiration)
		}
		// tenant service accounts may only access artifacts of their own namespace
		if namespace, isTenant := a.TenantServiceAccounts[username]; isTenant && c.QueryParam("namespace") != namespace {
			return echo.NewHTTPError(http.StatusForbidden, "service account not allowed to access this namespace")
		}
		c.Set("serviceAccount", username)
		return next(c)
	}
}

func (a *Authz) isServiceAccountAllowed(username string) bool {
	if _, ok := a.TenantServiceAccounts[username]; ok {
		return true
	}
	for _, sa := range a.ServiceAccounts {
		if sa == username {
			return true
		}
	}
	return false
}
