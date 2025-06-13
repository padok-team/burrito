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
	err := utils.InvalidateSession(c, sessionCookie)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to log out"})
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}

func HandleUserInfo(c echo.Context) error {
	// Check if the user is authenticated
	if c.Get("user_id") == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}
	userEmail := c.Get("user_email")
	name := c.Get("user_name")
	id := c.Get("user_id")
	picture := c.Get("user_picture")

	return c.JSON(http.StatusOK, map[string]string{"email": userEmail.(string), "name": name.(string), "id": id.(string), "picture": picture.(string)})
}
