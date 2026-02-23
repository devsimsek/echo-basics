package models

import (
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID        uuid.UUID `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
	Flag      FlagEnum  `gorm:"type:log_flag"`
	Message   string    `gorm:"type:text;not null"`
	Timestamp time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
