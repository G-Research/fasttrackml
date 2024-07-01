package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DefaultNamespaceCode represents default Namespace code.
const DefaultNamespaceCode = "default"

// Namespace represents a model to work with `namespaces` table.
type Namespace struct {
	ID                  uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Code                string         `gorm:"unique;index;not null" json:"code"`
	Description         string         `json:"description"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	DefaultExperimentID *int32         `gorm:"not null" json:"default_experiment_id"`
	Experiments         []Experiment   `gorm:"constraint:OnDelete:CASCADE" json:"experiments"`
}

// DisplayName returns Namespace display name.
func (ns Namespace) DisplayName() string {
	if ns.Description != "" {
		return fmt.Sprintf("%s (%s)", ns.Code, ns.Description)
	}
	return ns.Code
}

// IsDefault makes check that Namespace is default.
func (ns Namespace) IsDefault() bool {
	return ns.Code == DefaultNamespaceCode
}
