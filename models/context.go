package models

import "gorm.io/gorm"

// AppContext holds app-wide dependencies accessible in every handler
type AppContext struct {
	DB *gorm.DB
}
