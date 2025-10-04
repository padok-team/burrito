package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

func getStateGraphArgs(c echo.Context) (string, string, error) {
	namespace := c.QueryParam("namespace")
	layer := c.QueryParam("layer")
	if namespace == "" || layer == "" {
		return "", "", fmt.Errorf("missing query parameters")
	}
	return namespace, layer, nil
}

func (a *API) GetStateGraphHandler(c echo.Context) error {
	var err error
	var content []byte
	namespace, layer, err := getStateGraphArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	content, err = a.Storage.GetStateGraph(namespace, layer)
	if storageerrors.NotFound(err) {
		return c.String(http.StatusNotFound, "No state graph for this layer")
	}
	if err != nil {
		c.Logger().Errorf("Could not get state graph, there's an issue with the storage backend : %s", err)
		return c.String(http.StatusInternalServerError, "could not get state graph, there's an issue with the storage backend")
	}
	return c.JSONBlob(http.StatusOK, content)
}

func (a *API) PutStateGraphHandler(c echo.Context) error {
	var err error
	namespace, layer, err := getStateGraphArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	content, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not read request body: "+err.Error())
	}
	err = a.Storage.PutStateGraph(namespace, layer, content)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not put state graph, there's an issue with the storage backend: "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}
