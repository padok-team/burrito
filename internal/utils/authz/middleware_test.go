// nolint
package authz_test

import (
	"context"
	"path/filepath"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/utils/authz"
	v1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var cfg *rest.Config
var Client *client.Clientset
var testEnv *envtest.Environment
var Authz *authz.Authz
var e *echo.Echo

func TestAuthz(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Authz Middleware Suite")
}

func createServiceAccount(namespace string, name string) error {
	_, err := Client.CoreV1().ServiceAccounts(namespace).Create(context.TODO(), &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}, metav1.CreateOptions{})
	return err
}

func createToken(audience string, namespace string, name string) (string, error) {
	treq := &v1.TokenRequest{
		Spec: v1.TokenRequestSpec{
			Audiences: []string{audience},
		},
	}
	result, err := Client.CoreV1().ServiceAccounts(namespace).CreateToken(context.TODO(), name, treq, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return result.Status.Token, nil
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("../../..", "manifests", "crds")},
		ErrorIfCRDPathMissing: true,
	}
	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	Expect(err).NotTo(HaveOccurred())

	Authz = authz.NewAuthz()
	Expect(err).NotTo(HaveOccurred())
	Client, err = client.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Authz.Client = *Client
	Authz.SetAudience("datastore")
	Authz.AddServiceAccount("default", "authorized")
	createServiceAccount("default", "unauthorized")
	createServiceAccount("default", "authorized")
	e = echo.New()
})

var _ = Describe("Authz", func() {
	Describe("Nominal Case", Ordered, func() {
		var req *http.Request
		var rec *httptest.ResponseRecorder
		var context echo.Context
		var token string
		BeforeAll(func() {
			req = httptest.NewRequest(http.MethodGet, "/", nil)
			rec = httptest.NewRecorder()
			context = e.NewContext(req, rec)
		})
		Describe("When no Authorization header is present", Ordered, func() {
			It("should return unauthorized error", func() {
				err := Authz.Process(debugHandler)(context).(*echo.HTTPError)
				Expect(err.Code).To(Equal(401))
			})
		})
		Describe("When an Authorization header is present", Ordered, func() {
			It("should return 200 if the token is valid", func() {
				token, _ = createToken("datastore", "default", "authorized")
				req.Header.Set(echo.HeaderAuthorization, token)
				err := Authz.Process(debugHandler)(context)
				Expect(err).NotTo(HaveOccurred())
			})
			It("should return 200 again in cache", func() {
				req.Header.Set(echo.HeaderAuthorization, token)
				err := Authz.Process(debugHandler)(context)
				Expect(err).NotTo(HaveOccurred())
			})
			It("should return 401 if the token is from a non-allowed sa", func() {
				token, _ = createToken("datastore", "default", "unauthorized")
				req.Header.Set(echo.HeaderAuthorization, token)
				err := Authz.Process(debugHandler)(context).(*echo.HTTPError)
				Expect(err.Code).To(Equal(401))
			})
			It("should return 401 if the token is using a non-allowed audience", func() {
				token, _ = createToken("wrong", "default", "authorized")
				req.Header.Set(echo.HeaderAuthorization, token)
				err := Authz.Process(debugHandler)(context).(*echo.HTTPError)
				Expect(err.Code).To(Equal(401))
			})
		})
	})
})

func debugHandler(c echo.Context) error {
	return nil
}

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
