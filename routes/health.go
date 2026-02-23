package routes

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func HealthCheck(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Yeppers, seems good.",
	})
}
