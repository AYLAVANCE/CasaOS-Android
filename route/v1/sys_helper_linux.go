//go:build !android

package v1

import (
	"net/http"

	"github.com/IceWhaleTech/CasaOS/model"
	"github.com/IceWhaleTech/CasaOS/pkg/utils/common_err"
	"github.com/labstack/echo/v4"
)

func addExtraUtilization(data map[string]interface{}) {
}

// GetAndroidDeviceInfo returns Android device info placeholder on linux
func GetAndroidDeviceInfo(ctx echo.Context) error {
	return ctx.JSON(http.StatusNotImplemented, model.Result{Success: common_err.SERVICE_ERROR, Message: "Not implemented on non-Android platforms"})
}
