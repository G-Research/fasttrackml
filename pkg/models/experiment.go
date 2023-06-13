package models

import "database/sql"

// Experiment represents model to work with `experiments` table.
type Experiment struct {
	ID               *int32         `gorm:"column:experiment_id;not null;primaryKey"`
	Name             string         `gorm:"type:varchar(256);not null;unique"`
	ArtifactLocation string         `gorm:"type:varchar(256)"`
	LifecycleStage   LifecycleStage `gorm:"type:varchar(32);check:lifecycle_stage IN ('active', 'deleted')"`
	CreationTime     sql.NullInt64  `gorm:"type:bigint"`
	LastUpdateTime   sql.NullInt64  `gorm:"type:bigint"`
	Tags             []ExperimentTag
	Runs             []Run
	RunCount         *int32         `gorm:"type:bigint"`
}

// TableName returns actual table name.
func (o Experiment) TableName() string {
	return "experiments"
}

// ExperimentTag represents model to work with `experiment_tags` table.
type ExperimentTag struct {
	Key          string `gorm:"type:varchar(250);not null;primaryKey"`
	Value        string `gorm:"type:varchar(5000)"`
	ExperimentID int32  `gorm:"not null;primaryKey"`
}

// TableName returns actual table name.
func (o ExperimentTag) TableName() string {
	return "experiment_tags"
}
