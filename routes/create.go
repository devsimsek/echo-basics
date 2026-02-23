package routes

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"go.smsk.dev/pkgs/basics/echo-basics/models"
	"gorm.io/gorm"
)

func CreateLog(c *echo.Context) error {
	// task: make this atomic by utilising defer, DB.Begin() and commit.
	// get DB from app context
	db := c.Get("app").(*models.AppContext).DB
	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "app context missing"})
	}

	// request payload
	var payload struct {
		Flag    models.FlagEnum `json:"flag"`
		Message string          `json:"message" validate:"required"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "bad request, i had better expectations from you.",
		})
	}

	// Normalize flag to lowercase and trim whitespace
	flagStr := strings.ToLower(strings.TrimSpace(string(payload.Flag)))
	if flagStr == "" {
		// default to info if not provided
		flagStr = string(models.InfoFlag)
	}

	// Validate against allowed enum values
	allowed := map[string]struct{}{
		string(models.LogFlag):   {},
		string(models.DebugFlag): {},
		string(models.InfoFlag):  {},
		string(models.WarnFlag):  {},
		string(models.ErrorFlag): {},
		string(models.TraceFlag): {},
	}

	if _, ok := allowed[flagStr]; !ok {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "invalid flag value"})
	}

	// Build model
	log := models.Log{
		Flag:    models.FlagEnum(flagStr),
		Message: payload.Message,
	}

	// Use a transaction to make creation atomic
	if err := db.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, log)
}
