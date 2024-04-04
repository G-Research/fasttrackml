package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents model to work with `roles` table.
type Role struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role      string    `gorm:"unique;index;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
