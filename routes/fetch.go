package routes

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"go.smsk.dev/pkgs/basics/echo-basics/models"
	"gorm.io/gorm"
)

// FetchLogs returns all logs.
func FetchLogs(c *echo.Context) error {
	db := c.Get("app").(*models.AppContext).DB
	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "app context missing"})
	}

	var logs []models.Log
	if err := db.WithContext(c.Request().Context()).Find(&logs).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, logs)
}

// FetchID returns a single Log model by UUID.
// Returns 400 for bad id, 404 if not found, 500 for DB errors.
func FetchID(c *echo.Context) error {
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

	var log models.Log
	res := db.WithContext(c.Request().Context()).First(&log, "id = ?", id)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "ummm, wait a bit... Nope, its not here. Maybe you didn't send me that log?"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": res.Error.Error()})
	}

	return c.JSON(http.StatusOK, log)
}

// FetchTimestamp returns the latest log at or before the provided timestamp.
// Route param must be an RFC3339 timestamp string.
func FetchTimestamp(c *echo.Context) error {
	// task: make this atomic by utilising defer, DB.Begin() and commit.
	db := c.Get("app").(*models.AppContext).DB
	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "app context missing"})
	}

	tsStr := strings.TrimSpace(c.Param("timestamp"))
	if tsStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "bad request, i had better expectations from you."})
	}

	// Try RFC3339 and RFC3339Nano to be forgiving
	var ts time.Time
	var err error
	if ts, err = time.Parse(time.RFC3339, tsStr); err != nil {
		ts, err = time.Parse(time.RFC3339Nano, tsStr)
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "bad request, i had better expectations from you."})
	}

	var log models.Log
	res := db.WithContext(c.Request().Context()).
		Where("timestamp <= ?", ts).
		Order("timestamp desc").
		First(&log)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "ummm, wait a bit... Nope, its not here. Maybe you didn't send me that log?"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": res.Error.Error()})
	}

	return c.JSON(http.StatusOK, log)
}

// FetchFlag returns logs filtered by flag (case-insensitive). It validates the flag against allowed values.
func FetchFlag(c *echo.Context) error {
	// task: make this atomic by utilising defer, DB.Begin() and commit.
	db := c.Get("app").(*models.AppContext).DB
	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "app context missing"})
	}

	flagParam := strings.ToLower(strings.TrimSpace(c.Param("flag")))
	if flagParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "bad request, i had better expectations from you."})
	}

	allowed := map[string]struct{}{
		string(models.LogFlag):   {},
		string(models.DebugFlag): {},
		string(models.InfoFlag):  {},
		string(models.WarnFlag):  {},
		string(models.ErrorFlag): {},
		string(models.TraceFlag): {},
	}
	if _, ok := allowed[flagParam]; !ok {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "WHAT'S THAT FLAG? I HAVE NEVER SEEN THAT!!!"})
	}

	var logs []models.Log
	if err := db.WithContext(c.Request().Context()).Where("flag = ?", flagParam).Find(&logs).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, logs)
}
