package models

import (
	"time"

	"gorm.io/gorm"
)

// Namespace represents model to work with `namespaces` table.
type Namespace struct {
	ID                  uint   `gorm:"primaryKey;autoIncrement"`
	Code                string `gorm:"unique;index;not null"`
	Description         string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
	DefaultExperimentID *int32         `gorm:"not null"`
	Experiments         []Experiment   `gorm:"constraint:OnDelete:CASCADE"`
}
