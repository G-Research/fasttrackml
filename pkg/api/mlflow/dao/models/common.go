package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LifecycleStage represents entity stage
type LifecycleStage string

// Supported list of stages.
const (
	LifecycleStageActive  LifecycleStage = "active"
	LifecycleStageDeleted LifecycleStage = "deleted"
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
