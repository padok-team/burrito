package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func getStateGraphArgs(c echo.Context) (string, string, error) {
	namespace := c.Param("namespace")
	layer := c.Param("layer")
	if namespace == "" || layer == "" {
		return "", "", fmt.Errorf("missing query parameters")
	}
	return namespace, layer, nil
}

// stategraph/${namespace}/${layer}
func (a *API) GetStateGraphHandler(c echo.Context) error {
	namespace, layer, err := getStateGraphArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	content, err := a.Datastore.GetStateGraph(namespace, layer)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get state graph, there's an issue with the storage backend")
	}
	return c.JSONBlob(http.StatusOK, content)
}
