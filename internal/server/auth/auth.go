package auth

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type AuthHandlers interface {
	HandleLogin(c echo.Context) error
	HandleCallback(c echo.Context) error
}

func HandleLogout(c echo.Context, sessionCookie string) error {
	sess, err := session.Get(sessionCookie, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get session")
	}

	// Clear session
	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}
