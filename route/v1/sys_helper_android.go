//go:build android

package v1

import (
	"net/http"

	"github.com/IceWhaleTech/CasaOS/model"
	"github.com/IceWhaleTech/CasaOS/pkg/utils/common_err"
	"github.com/IceWhaleTech/CasaOS/service"
	"github.com/labstack/echo/v4"
)

func addExtraUtilization(data map[string]interface{}) {
	if bat, err := service.GetBatteryInfo(); err == nil {
		data["battery"] = bat
	}
	if wifi, err := service.GetWiFiInfo(); err == nil {
		data["wifi"] = wifi
	}
}

// GetAndroidDeviceInfo returns Android device info
func GetAndroidDeviceInfo(ctx echo.Context) error {
	info, err := service.GetAndroidDeviceInfo()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, model.Result{Success: common_err.SERVICE_ERROR, Message: err.Error()})
	}
	return ctx.JSON(http.StatusOK, model.Result{Success: common_err.SUCCESS, Message: common_err.GetMsg(common_err.SUCCESS), Data: info})
}
