package models

import (
	"time"

	"gorm.io/gorm"
)

// Namespace represents model to work with `namespaces` table.
type Namespace struct {
	ID                  uint           `gorm:"primaryKey;autoIncrement" json:"ID"`
	Code                string         `gorm:"unique;index;not null" json:"Code"`
	Description         string         `json:"Description"`
	CreatedAt           time.Time      `json:"CreatedAt"`
	UpdatedAt           time.Time      `json:"UpdatedAt"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"DeletedAt"`
	DefaultExperimentID *int32         `gorm:"not null" json:"DefaultExperimentID"`
	Experiments         []Experiment   `gorm:"constraint:OnDelete:CASCADE" json:"Experiments"`
}
