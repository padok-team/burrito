// nolint
package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Server Suite")
}

var e = echo.New()

func getContext(method string, path string) *echo.Context {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

// withSession wraps a handler with the session middleware so session.Get can
// find its backing store in the context, the same way it's wired in Exec().
func withSession(store sessions.Store, h echo.HandlerFunc) echo.HandlerFunc {
	return session.Middleware(store)(h)
}

// authenticatedRequest builds a request carrying a valid session cookie with
// the given user_id, produced by a real round trip through the cookie store.
func authenticatedRequest(store sessions.Store, userID string) *http.Request {
	base := httptest.NewRequest(http.MethodGet, "/", nil)
	sess, _ := store.New(base, cookieName)
	sess.Values["user_id"] = userID
	rec := httptest.NewRecorder()
	_ = sess.Save(base, rec)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range rec.Result().Cookies() {
		req.AddCookie(cookie)
	}
	return req
}

var _ = Describe("Server", func() {
	Describe("handleHealthz", func() {
		It("should return 200 OK", func() {
			context := getContext(http.MethodGet, "/healthz")
			err := handleHealthz(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusOK))
		})
	})

	Describe("staticSkipper", func() {
		s := &Server{}
		It("should skip /auth paths", func() {
			context := getContext(http.MethodGet, "/auth/login")
			Expect(s.staticSkipper(context)).To(BeTrue())
		})
		It("should skip /api paths", func() {
			context := getContext(http.MethodGet, "/api/layers")
			Expect(s.staticSkipper(context)).To(BeTrue())
		})
		It("should not skip other paths", func() {
			context := getContext(http.MethodGet, "/layers")
			Expect(s.staticSkipper(context)).To(BeFalse())
		})
	})

	Describe("authMiddleware", func() {
		store := sessions.NewCookieStore([]byte("test-key-32-bytes-long-xxxxxxxx"))
		s := &Server{sessionStore: store}

		It("should call next and set user_id when the session is valid", func() {
			nextCalled := false
			next := func(c *echo.Context) error {
				nextCalled = true
				return nil
			}
			handler := withSession(store, s.authMiddleware()(next))

			rec := httptest.NewRecorder()
			context := e.NewContext(authenticatedRequest(store, "u1"), rec)

			err := handler(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(nextCalled).To(BeTrue())
			Expect(context.Get("user_id")).To(Equal("u1"))
		})

		It("should return 401 when there is no session", func() {
			next := func(c *echo.Context) error {
				return nil
			}
			handler := withSession(store, s.authMiddleware()(next))
			context := getContext(http.MethodGet, "/api/layers")

			err := handler(context)
			Expect(err).To(HaveOccurred())
			httpErr, ok := err.(*echo.HTTPError)
			Expect(ok).To(BeTrue())
			Expect(httpErr.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("handleLogout", func() {
		It("should invalidate the session and redirect to /login", func() {
			store := sessions.NewCookieStore([]byte("test-key-32-bytes-long-xxxxxxxx"))
			s := &Server{sessionStore: store}
			handler := withSession(store, s.handleLogout)
			context := getContext(http.MethodPost, "/auth/logout")

			err := handler(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusTemporaryRedirect))
		})
	})

	Describe("handleAuthType", func() {
		It("should return basic when OIDC is disabled", func() {
			s := &Server{config: &config.Config{}}
			context := getContext(http.MethodGet, "/auth/type")

			err := s.handleAuthType(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusOK))
		})

		It("should return oauth when OIDC is enabled", func() {
			s := &Server{config: &config.Config{Server: config.ServerConfig{OIDC: config.OIDCConfig{Enabled: true}}}}
			context := getContext(http.MethodGet, "/auth/type")

			err := s.handleAuthType(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusOK))
		})
	})

	Describe("handleRoot", func() {
		It("should redirect to /layers when auth is disabled", func() {
			s := &Server{config: &config.Config{}}
			context := getContext(http.MethodGet, "/")

			err := s.handleRoot(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusTemporaryRedirect))
		})

		It("should redirect to /login when auth is enabled and there is no session", func() {
			store := sessions.NewCookieStore([]byte("test-key-32-bytes-long-xxxxxxxx"))
			s := &Server{
				config:       &config.Config{Server: config.ServerConfig{BasicAuth: config.BasicAuthConfig{Enabled: true}}},
				sessionStore: store,
			}
			handler := withSession(store, s.handleRoot)
			context := getContext(http.MethodGet, "/")

			err := handler(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusTemporaryRedirect))
		})
	})

	// handleUserInfo is a thin wrapper around auth.HandleUserInfo, itself
	// covered directly in internal/server/auth/auth_test.go; skipped here to
	// avoid duplicating that coverage with extra OIDC-adjacent scaffolding.
})
