package v1

import (
    "net/http"
    
    "github.com/IceWhaleTech/CasaOS/model"
    "github.com/labstack/echo/v4"
)

// RequireRole returns middleware that restricts access based on user role
func RequireRole(requiredRoles ...model.UserRole) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Get role from JWT claims (stored in context by auth middleware)
            roleClaim := c.Get("user_role")
            if roleClaim == nil {
                return c.JSON(http.StatusForbidden, model.Result{
                    Success: http.StatusForbidden,
                    Message: "Access denied: no role assigned",
                })
            }
            
            userRole := model.UserRole(roleClaim.(string))
            for _, required := range requiredRoles {
                if userRole == required {
                    return next(c)
                }
            }
            
            return c.JSON(http.StatusForbidden, model.Result{
                Success: http.StatusForbidden,
                Message: "Access denied: insufficient permissions",
            })
        }
    }
}

// RequirePermission checks a specific permission
func RequirePermission(check func(model.Permission) bool) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            roleClaim := c.Get("user_role")
            if roleClaim == nil {
                return c.JSON(http.StatusForbidden, model.Result{
                    Success: http.StatusForbidden,
                    Message: "Access denied",
                })
            }
            
            role := model.UserRole(roleClaim.(string))
            perms := model.GetPermissionsForRole(role)
            if !check(perms) {
                return c.JSON(http.StatusForbidden, model.Result{
                    Success: http.StatusForbidden,
                    Message: "Access denied: insufficient permissions for this resource",
                })
            }
            return next(c)
        }
    }
}
