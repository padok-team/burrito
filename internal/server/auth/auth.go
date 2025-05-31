package auth

import (
	"net/http"

	"github.com/padok-team/burrito/internal/server/utils"

	"github.com/labstack/echo/v4"
)

type AuthHandlers interface {
	HandleLogin(c echo.Context) error
	HandleCallback(c echo.Context) error
	GetLoginHTTPMethod() string
}

func HandleLogout(c echo.Context, sessionCookie string) error {
	utils.InvalidateSession(c, sessionCookie)

	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}
