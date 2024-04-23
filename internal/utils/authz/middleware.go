package authz

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
)

type (
	Authz struct {
		Cache           *cache.Cache
		Client          client.Clientset
		Audience        string
		ServiceAccounts []string
	}
)

func NewAuthz() *Authz {
	return &Authz{
		Cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

func (a *Authz) SetAudience(audience string) {
	a.Audience = audience
}

func (a *Authz) AddServiceAccount(namespace string, name string) {
	a.ServiceAccounts = append(a.ServiceAccounts, fmt.Sprintf("system:serviceaccount:%s:%s", namespace, name))
}

// Process is the middleware function.
func (a *Authz) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authz := c.Request().Header.Get("Authorization")
		if authz == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing Authorization header")
		}
		_, found := a.Cache.Get(authz)
		if found {
			return next(c)
		}
		ctx := context.Background()
		tr := &v1.TokenReview{
			Spec: v1.TokenReviewSpec{
				Token:     authz,
				Audiences: []string{a.Audience},
			},
		}
		result, err := a.Client.AuthenticationV1().TokenReviews().Create(ctx, tr, metav1.CreateOptions{})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "error validating token")
		}
		if !result.Status.Authenticated {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}
		fmt.Printf("%s", result.Status.User.Username)
		if !a.isServiceAccountAllowed(result.Status.User.Username) {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized user")
		}
		a.Cache.Set(authz, true, cache.DefaultExpiration)
		return next(c)
	}
}

func (a *Authz) isServiceAccountAllowed(username string) bool {
	for _, sa := range a.ServiceAccounts {
		if sa == username {
			return true
		}
	}
	return false
}
