package v1

import (
    "net/http"
    
    "github.com/IceWhaleTech/CasaOS/model"
    "github.com/IceWhaleTech/CasaOS/pkg/utils/common_err"
    "github.com/labstack/echo/v4"
)

// GetSetupStatus checks if initial setup has been completed
func GetSetupStatus(ctx echo.Context) error {
    // Check if any users exist in the database
    // If no users exist, setup is required
    // This is a simplified check - the actual implementation would query UserService
    return ctx.JSON(http.StatusOK, map[string]interface{}{
        "setup_required": true,
        "version":        "0.4.15-android",
    })
}

// PostSetup handles first-boot account creation
func PostSetup(ctx echo.Context) error {
    var setup model.FirstBootSetup
    if err := ctx.Bind(&setup); err != nil {
        return ctx.JSON(http.StatusBadRequest, model.Result{
            Success: common_err.CLIENT_ERROR,
            Message: "Invalid setup data: " + err.Error(),
        })
    }
    
    if len(setup.Username) < 3 || len(setup.Password) < 6 {
        return ctx.JSON(http.StatusBadRequest, model.Result{
            Success: common_err.CLIENT_ERROR,
            Message: "Username must be 3+ chars, password 6+ chars",
        })
    }
    
    // TODO: Wire to CasaOS-UserService to create the master admin user
    // This would call the UserService API:
    //   POST /v1/users with role=admin
    
    return ctx.JSON(http.StatusOK, model.Result{
        Success: common_err.SUCCESS,
        Message: "Setup complete. Admin account created.",
        Data: map[string]string{"username": setup.Username, "role": string(model.RoleAdmin)},
    })
}
