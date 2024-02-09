package models

import (
	"database/sql"
)

// Default Experiment properties.
const (
	DefaultExperimentID   = int32(0)
	DefaultExperimentName = "Default"
)

// Experiment represents model to work with `experiments` table.
type Experiment struct {
	ID               *int32         `gorm:"column:experiment_id;not null;primaryKey"`
	Name             string         `gorm:"type:varchar(256);not null;index:,unique,composite:name"`
	ArtifactLocation string         `gorm:"type:varchar(256)"`
	LifecycleStage   LifecycleStage `gorm:"type:varchar(32);check:lifecycle_stage IN ('active', 'deleted')"`
	CreationTime     sql.NullInt64  `gorm:"type:bigint"`
	LastUpdateTime   sql.NullInt64  `gorm:"type:bigint"`
	NamespaceID      uint           `gorm:"index:,unique,composite:name"`
	Namespace        Namespace
	Tags             []ExperimentTag `gorm:"constraint:OnDelete:CASCADE"`
	Runs             []Run           `gorm:"constraint:OnDelete:CASCADE"`
}

// IsDefault makes check that Experiment is default.
func (e Experiment) IsDefault() bool {
	return e.ID != nil && *e.ID == DefaultExperimentID && e.Name == DefaultExperimentName
}

// ExperimentTag represents model to work with `experiment_tags` table.
type ExperimentTag struct {
	Key          string `gorm:"type:varchar(250);not null;primaryKey"`
	Value        string `gorm:"type:varchar(5000)"`
	ExperimentID int32  `gorm:"not null;primaryKey"`
}
