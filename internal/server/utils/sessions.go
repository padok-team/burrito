package utils

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// RemoveSessionCookie manually removes the session cookie from the response.
// This is useful when the session does not exist server-side, but client-side cookies still exist
func RemoveSessionCookie(c echo.Context, sessionCookie string) error {
	http.SetCookie(c.Response(), &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.Request().TLS != nil,
	})
	return nil
}

// InvalidateSession clears the session data and sets the session cookie to expire immediately.
func InvalidateSession(c echo.Context, sessionCookie string) error {
	sess, err := session.Get(sessionCookie, c)
	if err != nil {
		log.Warn("Tried to invalidate session, but session was not found")
	}

	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}
	return nil
}
