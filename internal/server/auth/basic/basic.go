package basic

import (
	"context"
	"fmt"
	"net/http"

	"crypto/rand"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/server/utils"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultSecretName = "burrito-admin-credentials"
	DefaultUsername   = "admin"
)

type BasicAuthHandlers struct {
	Username        string
	Password        string
	SessionCookie   string
	LoginHTTPMethod string
}

func New(c *config.Config, ctx context.Context, cl client.Client, sessionCookie string) (*BasicAuthHandlers, error) {
	secret := &corev1.Secret{}
	err := cl.Get(ctx, types.NamespacedName{
		Name:      DefaultSecretName,
		Namespace: c.Controller.MainNamespace,
	}, secret)
	// if secret does not exist, generate a new one with "admin" as username and a random password
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, err
		}
		log.Info("No existing secret found for basic auth, generating a new one...")
		// Generate a new secret with "admin" as username and a random password
		defaultPassword := rand.Text()
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DefaultSecretName,
				Namespace: c.Controller.MainNamespace,
			},
			Type: corev1.SecretTypeBasicAuth,
			Data: map[string][]byte{
				"username": []byte(DefaultUsername),
				"password": []byte(defaultPassword),
			},
		}
		err = cl.Create(ctx, secret)
		if err != nil {
			return nil, fmt.Errorf("error creating basic auth credentials secret: %v", err)
		}
	}
	return &BasicAuthHandlers{
		Username:        string(secret.Data["username"]),
		Password:        string(secret.Data["password"]),
		SessionCookie:   sessionCookie,
		LoginHTTPMethod: http.MethodPost,
	}, nil
}

func (b *BasicAuthHandlers) HandleLogin(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	if username == b.Username && password == b.Password {
		// Set session cookie
		sess, err := session.Get(b.SessionCookie, c)
		if err != nil {
			// Clear session cookie if session is invalid to prevent stale sessions
			err := utils.RemoveSessionCookie(c, b.SessionCookie)
			if err != nil {
				log.Warnf("Failed to clear session cookie: %v", err)
			}
		}

		// Upgrade session cookie to SameSite=Strict after successful login
		sess.Options.SameSite = http.SameSiteStrictMode

		sess.Values["user_id"] = "admin"
		sess.Values["email"] = "admin@burrito.tf"
		sess.Values["name"] = "admin"
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
		}
	} else {
		log.Warnf("Failed login attempt with username: %s", username)
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid username or password")
	}
	return nil
}

func (b *BasicAuthHandlers) GetLoginHTTPMethod() string {
	return b.LoginHTTPMethod
}

func (b *BasicAuthHandlers) HandleCallback(c echo.Context) error {
	// Basic auth does not require a callback, so we can just redirect to the home page
	return c.Redirect(http.StatusFound, "/")
}
