//go:build android

package v1

import (
	"net/http"

	"github.com/IceWhaleTech/CasaOS/service"
	"github.com/labstack/echo/v4"
)

var ls *service.LocalStorageAndroid

func init() {
	ls = service.NewLocalStorageAndroid()
}

func GetStorageVolumes(ctx echo.Context) error {
	volumes, err := ls.GetVolumes()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, volumes)
}

func GetStorageVolumeByPath(ctx echo.Context) error {
	path := ctx.Param("path")
	volumes, err := ls.GetVolumes()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	for _, v := range volumes {
		if v.Path == path {
			return ctx.JSON(http.StatusOK, v)
		}
	}
	return ctx.JSON(http.StatusNotFound, map[string]string{"error": "volume not found"})
}

func GetStorageHealth(ctx echo.Context) error {
	health := ls.GetStorageHealth()
	return ctx.JSON(http.StatusOK, health)
}

func GetStorageSAFFolders(ctx echo.Context) error {
	// Dummy for Android
	return ctx.JSON(http.StatusOK, []string{})
}

func PostStorageSAFRequest(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "Not implemented"})
}

func GetStorageDetect(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "detected"})
}

func GetStorageUsage(ctx echo.Context) error {
	usage, err := service.CalculateUsage()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, usage)
}
