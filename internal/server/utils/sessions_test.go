// nolint
package utils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/server/utils"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Server Utils Suite")
}

const testSessionCookie = "test_session"

func getContext(method string, path string) *echo.Context {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

var _ = Describe("RemoveSessionCookie", func() {
	It("should set an expired cookie on the response", func() {
		context := getContext(http.MethodPost, "/auth/logout")

		err := utils.RemoveSessionCookie(context, testSessionCookie)
		Expect(err).NotTo(HaveOccurred())

		cookies := context.Response().(*echo.Response).Header().Values("Set-Cookie")
		Expect(cookies).NotTo(BeEmpty())
	})
})

var _ = Describe("InvalidateSession", func() {
	It("should clear the session and save an expired cookie", func() {
		store := sessions.NewCookieStore([]byte("test-key-32-bytes-long-xxxxxxxx"))
		context := getContext(http.MethodPost, "/auth/logout")

		handler := session.Middleware(store)(func(c *echo.Context) error {
			return utils.InvalidateSession(c, testSessionCookie)
		})

		err := handler(context)
		Expect(err).NotTo(HaveOccurred())
	})
})
