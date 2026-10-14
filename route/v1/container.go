package v1

import (
	"fmt"
	"net/http"

	"github.com/IceWhaleTech/CasaOS/pkg/container"
	"github.com/labstack/echo/v4"
)

var containerManager = container.NewManager()

// GetContainers returns a list of all containers
func GetContainers(c echo.Context) error {
	rt := containerManager.GetActive()
	containers, err := rt.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, containers)
}

// CreateContainer creates and starts a new container
func CreateContainer(c echo.Context) error {
	var req container.ContainerCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	
	rt := containerManager.GetActive()
	id, err := rt.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, map[string]string{"id": id, "message": "Container created successfully"})
}

func InspectContainer(c echo.Context) error {
	id := c.Param("id")
	rt := containerManager.GetActive()
	
	container, err := rt.Inspect(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, container)
}

func StartContainer(c echo.Context) error {
	id := c.Param("id")
	rt := containerManager.GetActive()
	
	if err := rt.Start(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, map[string]string{"message": "Container started"})
}

func StopContainer(c echo.Context) error {
	id := c.Param("id")
	rt := containerManager.GetActive()
	
	if err := rt.Stop(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, map[string]string{"message": "Container stopped"})
}

func RemoveContainer(c echo.Context) error {
	id := c.Param("id")
	force := c.QueryParam("force") == "true"
	rt := containerManager.GetActive()
	
	if err := rt.Remove(c.Request().Context(), id, force); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, map[string]string{"message": "Container removed"})
}

func GetImages(c echo.Context) error {
	rt := containerManager.GetActive()
	images, err := rt.ListImages(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, images)
}

func PullImage(c echo.Context) error {
	var body struct {
		Image string `json:"image"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	
	rt := containerManager.GetActive()
	if err := rt.PullImage(c.Request().Context(), body.Image); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Image %s pulled successfully", body.Image)})
}

func RemoveImage(c echo.Context) error {
	id := c.Param("id")
	rt := containerManager.GetActive()
	
	if err := rt.RemoveImage(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, map[string]string{"message": "Image removed"})
}

func GetRuntimeInfo(c echo.Context) error {
	rt := containerManager.GetActive()
	version, _ := rt.Version()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"name": rt.Name(),
		"version": version,
		"available": rt.IsAvailable(),
	})
}
