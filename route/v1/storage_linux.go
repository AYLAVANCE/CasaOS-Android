//go:build !android

package v1

import (
	"net/http"

	"github.com/IceWhaleTech/CasaOS/service"
	"github.com/labstack/echo/v4"
)

func GetStorageVolumes(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, []interface{}{})
}

func GetStorageVolumeByPath(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotFound, map[string]string{"error": "storage volumes only available on Android"})
}

func GetStorageHealth(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "native"})
}

func GetStorageSAFFolders(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, []string{})
}

func PostStorageSAFRequest(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, map[string]string{"error": "SAF only available on Android"})
}

func GetStorageDetect(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": "not applicable"})
}

func GetStorageUsage(ctx echo.Context) error {
	usage, err := service.CalculateUsage()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, usage)
}
