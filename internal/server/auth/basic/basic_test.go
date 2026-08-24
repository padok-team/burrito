// nolint
package basic_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/server/auth/basic"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBasicAuth(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Basic Auth Suite")
}

const testSessionCookie = "test_session"

func getContext(method string, path string, body string) *echo.Context {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

var _ = Describe("basic.New", func() {
	It("should generate a new secret when none exists and return valid handlers", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		cl := fake.NewClientBuilder().WithScheme(scheme).Build()
		cfg := &config.Config{Controller: config.ControllerConfig{MainNamespace: "burrito-system"}}

		handlers, err := basic.New(cfg, context.Background(), cl, testSessionCookie)
		Expect(err).NotTo(HaveOccurred())
		Expect(handlers.Username).To(Equal(basic.DefaultUsername))
		Expect(handlers.Password).NotTo(BeEmpty())
	})

	It("should reuse an existing secret", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      basic.DefaultSecretName,
				Namespace: "burrito-system",
			},
			Type: corev1.SecretTypeBasicAuth,
			Data: map[string][]byte{
				"username": []byte("testuser"),
				"password": []byte("testpass"),
			},
		}
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()
		cfg := &config.Config{Controller: config.ControllerConfig{MainNamespace: "burrito-system"}}

		handlers, err := basic.New(cfg, context.Background(), cl, testSessionCookie)
		Expect(err).NotTo(HaveOccurred())
		Expect(handlers.Username).To(Equal("testuser"))
		Expect(handlers.Password).To(Equal("testpass"))
	})
})

var _ = Describe("BasicAuthHandlers", func() {
	handlers := &basic.BasicAuthHandlers{
		Username:        "testuser",
		Password:        "testpass",
		SessionCookie:   testSessionCookie,
		LoginHTTPMethod: http.MethodPost,
	}

	Describe("HandleLogin", func() {
		It("should return 401 for invalid credentials", func() {
			form := url.Values{"username": {"testuser"}, "password": {"wrong"}}
			context := getContext(http.MethodPost, "/auth/login", form.Encode())

			err := handlers.HandleLogin(context)
			Expect(err).To(HaveOccurred())
			httpErr, ok := err.(*echo.HTTPError)
			Expect(ok).To(BeTrue())
			Expect(httpErr.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should set the session and return nil for valid credentials", func() {
			store := sessions.NewCookieStore([]byte("test-key-32-bytes-long-xxxxxxxx"))
			form := url.Values{"username": {"testuser"}, "password": {"testpass"}}
			context := getContext(http.MethodPost, "/auth/login", form.Encode())

			handler := session.Middleware(store)(handlers.HandleLogin)
			err := handler(context)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetLoginHTTPMethod", func() {
		It("should return the configured HTTP method", func() {
			Expect(handlers.GetLoginHTTPMethod()).To(Equal(http.MethodPost))
		})
	})

	Describe("HandleCallback", func() {
		It("should redirect to /", func() {
			context := getContext(http.MethodGet, "/auth/callback", "")

			err := handlers.HandleCallback(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusFound))
		})
	})
})
