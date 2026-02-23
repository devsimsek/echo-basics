package routes

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"go.smsk.dev/pkgs/basics/echo-basics/models"
	"go.smsk.dev/pkgs/basics/echo-basics/modules"
	"gorm.io/gorm"
)

func DeleteLog(c *echo.Context) error {
	// task: make this atomic by utilising defer, DB.Begin() and commit.
	// get DB from app context
	db := c.Get("app").(*models.AppContext).DB
	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "app context missing"})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "bad request, i had better expectations from you.",
		})
	}

	var log models.Log = models.Log{
		ID: id,
	}

	res := db.WithContext(c.Request().Context()).Find(&log)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "ummm, wait a bit... Nope, its not here. Maybe you didn't send me that log?"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": res.Error.Error()})
	}

	if modules.GetLogLevel(log.Flag) < 4 {

		res = db.WithContext(c.Request().Context()).Delete(&log)

		return c.JSON(http.StatusOK, map[string]string{"message": "That record is long gone now. Don't worry, our secret is now safe."})
	}

	return c.JSON(http.StatusForbidden, map[string]string{"error": "EEEEYYYY! You can't do that!"})
}
