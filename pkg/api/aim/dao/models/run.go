package models

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Status represents Status type.
type Status string

// Supported list of statuses.
const (
	StatusRunning   Status = "RUNNING"
	StatusScheduled Status = "SCHEDULED"
	StatusFinished  Status = "FINISHED"
	StatusFailed    Status = "FAILED"
	StatusKilled    Status = "KILLED"
)

// Run represents model to work with `runs` table.
//
//nolint:lll
type Run struct {
	ID             string         `gorm:"<-:create;column:run_uuid;type:varchar(32);not null;primaryKey"`
	Name           string         `gorm:"type:varchar(250)"`
	SourceType     string         `gorm:"<-:create;type:varchar(20);check:source_type IN ('NOTEBOOK', 'JOB', 'LOCAL', 'UNKNOWN', 'PROJECT')"`
	SourceName     string         `gorm:"<-:create;type:varchar(500)"`
	EntryPointName string         `gorm:"<-:create;type:varchar(50)"`
	UserID         string         `gorm:"<-:create;type:varchar(256)"`
	Status         Status         `gorm:"type:varchar(9);check:status IN ('SCHEDULED', 'FAILED', 'FINISHED', 'RUNNING', 'KILLED')"`
	StartTime      sql.NullInt64  `gorm:"<-:create;type:bigint"`
	EndTime        sql.NullInt64  `gorm:"type:bigint"`
	SourceVersion  string         `gorm:"<-:create;type:varchar(50)"`
	LifecycleStage LifecycleStage `gorm:"type:varchar(20);check:lifecycle_stage IN ('active', 'deleted')"`
	ArtifactURI    string         `gorm:"<-:create;type:varchar(200)"`
	ExperimentID   int32
	Experiment     Experiment
	DeletedTime    sql.NullInt64  `gorm:"type:bigint"`
	RowNum         RowNum         `gorm:"<-:create;index"`
	Params         []Param        `gorm:"constraint:OnDelete:CASCADE"`
	Tags           []Tag          `gorm:"constraint:OnDelete:CASCADE"`
	SharedTags     []SharedTag    `gorm:"many2many:run_shared_tags"`
	Logs           []Log          `gorm:"constraint:OnDelete:CASCADE"`
	Metrics        []Metric       `gorm:"constraint:OnDelete:CASCADE"`
	LatestMetrics  []LatestMetric `gorm:"constraint:OnDelete:CASCADE"`
}

// RowNum represents custom data type.
type RowNum int64

// Scan implements Gorm interface for custom data types.
func (rn *RowNum) Scan(v interface{}) error {
	nullInt := sql.NullInt64{}
	if err := nullInt.Scan(v); err != nil {
		return err
	}
	*rn = RowNum(nullInt.Int64)
	return nil
}

// GormDataType implements Gorm interface for custom data types.
func (rn RowNum) GormDataType() string {
	return "bigint"
}

// GormValue implements Gorm interface for custom data types.
func (rn RowNum) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if rn == 0 {
		return clause.Expr{
			SQL: "(SELECT COALESCE(MAX(row_num), -1) FROM runs) + 1",
		}
	}
	return clause.Expr{
		SQL:  "?",
		Vars: []interface{}{int64(rn)},
	}
}
