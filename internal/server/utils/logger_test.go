// nolint
package utils_test

import (
	"net/http"

	"github.com/labstack/echo/v5/middleware"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/server/utils"
)

var _ = Describe("LoggerMiddlewareConfig.LogValuesFunc", func() {
	It("should log the request and return nil", func() {
		context := getContext(http.MethodGet, "/api/layers")
		context.Set("user_email", "user@example.com")

		err := utils.LoggerMiddlewareConfig.LogValuesFunc(context, middleware.RequestLoggerValues{
			RemoteIP: "127.0.0.1",
			Host:     "localhost",
			Method:   http.MethodGet,
			URI:      "/api/layers",
			Status:   http.StatusOK,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should log unauthenticated when no user_email is set", func() {
		context := getContext(http.MethodGet, "/api/layers")

		err := utils.LoggerMiddlewareConfig.LogValuesFunc(context, middleware.RequestLoggerValues{})
		Expect(err).NotTo(HaveOccurred())
	})
})
