package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

func getRevisionArgs(c echo.Context) (string, string, string, error) {
	namespace := c.QueryParam("namespace")
	name := c.QueryParam("name")
	ref := c.QueryParam("ref")
	if namespace == "" || name == "" || ref == "" {
		return "", "", "", fmt.Errorf("missing query parameters")
	}
	return namespace, name, ref, nil
}

func (a *API) PutGitBundleHandler(c echo.Context) error {
	namespace, name, ref, err := getRevisionArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	revision := c.QueryParam("revision")
	if revision == "" {
		return c.String(http.StatusBadRequest, "missing revision parameter")
	}

	content, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not read request body: "+err.Error())
	}

	err = a.Storage.PutGitBundle(namespace, name, ref, revision, content)
	if err != nil {
		c.Logger().Errorf("Could not store revision, there's an issue with the storage backend: %s", err)
		return c.String(http.StatusInternalServerError, "could not store revision, there's an issue with the storage backend")
	}

	return c.NoContent(http.StatusOK)
}

func (a *API) HeadGitBundleHandler(c echo.Context) error {
	namespace, name, ref, err := getRevisionArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	revision := c.QueryParam("revision")
	if revision == "" {
		return c.String(http.StatusBadRequest, "missing revision parameter")
	}
	checksum, err := a.Storage.CheckGitBundle(namespace, name, ref, revision)
	if err != nil {
		if storageerrors.NotFound(err) {
			return c.String(http.StatusNotFound, "No bundle found for this revision")
		}
		c.Logger().Errorf("Could not get bundle for revision, there's an issue with the storage backend: %s", err)
		return c.String(http.StatusInternalServerError, "could not get bundle for revision, there's an issue with the storage backend")
	}

	return c.String(http.StatusOK, string(checksum))
}

func (a *API) GetGitBundleHandler(c echo.Context) error {
	namespace, name, ref, err := getRevisionArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	revision := c.QueryParam("revision")
	if revision == "" {
		return c.String(http.StatusBadRequest, "missing revision parameter")
	}
	content, err := a.Storage.GetGitBundle(namespace, name, ref, revision)
	if err != nil {
		if storageerrors.NotFound(err) {
			return c.String(http.StatusNotFound, "No bundle found for this revision")
		}
		c.Logger().Errorf("Could not get bundle for revision, there's an issue with the storage backend: %s", err)
		return c.String(http.StatusInternalServerError, "could not get bundle for revision, there's an issue with the storage backend")
	}

	return c.Blob(http.StatusOK, "application/octet-stream", content)
}
