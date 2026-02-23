package modules

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDB opens a database connection and returns a *gorm.DB.
// It intentionally does NOT perform AutoMigrate or any schema changes
// so that opening the DB does not alter the database schema.
func InitDB() *gorm.DB {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		panic("DSN environment variable is not set")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // this disables implicit prepared statement usage. set false for production if needed.
	}), &gorm.Config{})

	if err != nil {
		panic("There exist a connection error, check logs manually: " + err.Error())
	}

	return db
}