package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

// Base is a base model which holds common fields for each model.
type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate triggers by GORM before create.
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	b.ID = uuid.New()
	return nil
}
