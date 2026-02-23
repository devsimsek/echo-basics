package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"go.smsk.dev/pkgs/basics/echo-basics/models"
	"gorm.io/gorm"
)

// Run applies migrations in an idempotent way.
// It will create the uuid extension, the log_flag enum type (if missing)
// and the logs table (if missing). Rollback will drop the table and the enum only if they exist.
func Run(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "1771799054_init_uuid_and_logs",
			Migrate: func(tx *gorm.DB) error {
				// Ensure uuid extension exists (Postgres)
				if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
					return err
				}

				// Create enum type only if it doesn't exist
				var typCount int64
				if err := tx.Raw(`SELECT count(*) FROM pg_type WHERE typname = ?`, "log_flag").Scan(&typCount).Error; err != nil {
					return err
				}
				if typCount == 0 {
					if err := tx.Exec(`CREATE TYPE log_flag AS ENUM ('log', 'debug', 'info', 'warn', 'error', 'trace');`).Error; err != nil {
						return err
					}
				}

				// Create logs table only if it doesn't exist
				hasTable := tx.Migrator().HasTable(&models.Log{})
				if !hasTable {
					if err := tx.Migrator().CreateTable(&models.Log{}); err != nil {
						return err
					}
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				// Drop logs table if exists
				if tx.Migrator().HasTable(&models.Log{}) {
					if err := tx.Migrator().DropTable(&models.Log{}); err != nil {
						return err
					}
				}

				// Drop enum type if exists (use DO block to conditionally drop)
				if err := tx.Exec(`
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'log_flag') THEN
    DROP TYPE log_flag;
  END IF;
END
$$;`).Error; err != nil {
					return err
				}

				return nil
			},
		},
	})

	return m.Migrate()
}
