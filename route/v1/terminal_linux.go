//go:build !android

package v1

import (
	"github.com/labstack/echo/v4"
)

// HandleWebTerminal is a no-op or handles linux specific terminal behavior
func HandleWebTerminal(ctx echo.Context) error {
	return nil
}
