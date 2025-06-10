package utils

import (
	log "github.com/sirupsen/logrus"

	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var LoggerMiddlewareConfig = middleware.RequestLoggerConfig{
	LogLatency:       true,
	LogRemoteIP:      true,
	LogHost:          true,
	LogMethod:        true,
	LogURI:           true,
	LogStatus:        true,
	LogError:         true,
	LogResponseSize:  true,
	LogContentLength: true,
	LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
		email := "unauthenticated"
		if e := c.Get("user_email"); e != nil {
			email, _ = e.(string)
		}
		log.WithFields(log.Fields{
			"time":          time.Now().Format(time.RFC3339Nano),
			"remote_ip":     v.RemoteIP,
			"host":          v.Host,
			"method":        v.Method,
			"uri":           v.URI,
			"status":        v.Status,
			"error":         v.Error,
			"latency":       v.Latency.Nanoseconds() / 1000,
			"latency_human": v.Latency,
			"bytes_in":      v.ContentLength,
			"bytes_out":     v.ResponseSize,
			"email":         email,
		}).Info("request")
		return nil
	},
}
