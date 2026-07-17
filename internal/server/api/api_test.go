package api_test

import (
	"testing"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestServerAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Server API Suite")
}

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = configv1alpha1.AddToScheme(s)
	return s
}

func setRouteParams(c echo.Context, names []string, values []string) {
	c.SetParamNames(names...)
	c.SetParamValues(values...)
}

func boolPtr(b bool) *bool {
	return &b
}
