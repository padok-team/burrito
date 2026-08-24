// nolint
package webhook_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/repository/credentials"
	"github.com/padok-team/burrito/internal/webhook"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWebhook(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Suite")
}

var _ = Describe("Webhook", func() {
	Describe("GetHttpHandler", func() {
		It("should process a webhook request through the mock provider", func() {
			scheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			Expect(configv1alpha1.AddToScheme(scheme)).To(Succeed())

			repository := &configv1alpha1.TerraformRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "repo",
					Namespace: "default",
				},
				Spec: configv1alpha1.TerraformRepositorySpec{
					Repository: configv1alpha1.TerraformRepositoryRepository{
						Url: "https://git.mock.com/repo",
					},
				},
			}
			credentialSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "repository-secret",
					Namespace: "default",
				},
				Type: credentials.CredentialsType,
				Data: map[string][]byte{
					"provider":      []byte("mock"),
					"url":           []byte(repository.Spec.Repository.Url),
					"webhookSecret": []byte("s3cr3t"),
				},
			}

			cl := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(repository, credentialSecret).
				WithIndex(&corev1.Secret{}, "type", func(obj kclient.Object) []string {
					secret := obj.(*corev1.Secret)
					return []string{string(secret.Type)}
				}).
				Build()

			w := webhook.New(&config.Config{
				Controller: config.ControllerConfig{
					Timers: config.ControllerTimers{
						CredentialsTTL: time.Hour,
					},
				},
			}, cl)

			handler := w.GetHttpHandler()
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/webhook", strings.NewReader("{}"))
			rec := httptest.NewRecorder()
			context := e.NewContext(req, rec)

			err := handler(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).NotTo(Equal(0))
		})
	})
})
