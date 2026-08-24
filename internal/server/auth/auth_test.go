// nolint
package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/server/auth"
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Auth Suite")
}

const testSessionCookie = "test_session"

func getContext(method string, path string) *echo.Context {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

var _ = Describe("HandleLogout", func() {
	It("should invalidate the session and redirect to /login", func() {
		store := sessions.NewCookieStore([]byte("test-key-32-bytes-long-xxxxxxxx"))
		context := getContext(http.MethodPost, "/auth/logout")

		handler := session.Middleware(store)(func(c *echo.Context) error {
			return auth.HandleLogout(c, testSessionCookie)
		})

		err := handler(context)
		Expect(err).NotTo(HaveOccurred())
		Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusTemporaryRedirect))
	})
})

var _ = Describe("HandleUserInfo", func() {
	It("should return 401 when the user is not authenticated", func() {
		context := getContext(http.MethodGet, "/auth/user")
		err := auth.HandleUserInfo(context)
		Expect(err).NotTo(HaveOccurred())
		Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusUnauthorized))
	})

	It("should return the user info when authenticated", func() {
		context := getContext(http.MethodGet, "/auth/user")
		context.Set("user_id", "u1")
		context.Set("user_email", "user@example.com")
		context.Set("user_name", "User Name")
		context.Set("user_picture", "https://example.com/avatar.png")

		err := auth.HandleUserInfo(context)
		Expect(err).NotTo(HaveOccurred())
		Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusOK))
	})
})
