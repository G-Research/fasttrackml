package models

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Run struct {
	ID             string         `gorm:"column:run_uuid;type:varchar(32);not null;primaryKey"`
	Name           string         `gorm:"type:varchar(250)"`
	SourceType     string         `gorm:"type:varchar(20);check:source_type IN ('NOTEBOOK', 'JOB', 'LOCAL', 'UNKNOWN', 'PROJECT')"`
	SourceName     string         `gorm:"type:varchar(500)"`
	EntryPointName string         `gorm:"type:varchar(50)"`
	UserID         string         `gorm:"type:varchar(256)"`
	Status         Status         `gorm:"type:varchar(9);check:status IN ('SCHEDULED', 'FAILED', 'FINISHED', 'RUNNING', 'KILLED')"`
	StartTime      sql.NullInt64  `gorm:"type:bigint"`
	EndTime        sql.NullInt64  `gorm:"type:bigint"`
	SourceVersion  string         `gorm:"type:varchar(50)"`
	LifecycleStage LifecycleStage `gorm:"type:varchar(20);check:lifecycle_stage IN ('active', 'deleted')"`
	ArtifactURI    string         `gorm:"type:varchar(200)"`
	ExperimentID   int32
	Experiment     Experiment
	DeletedTime    sql.NullInt64 `gorm:"type:bigint"`
	RowNum         RowNum        `gorm:"index"`
	Params         []Param
	Tags           []Tag
	Metrics        []Metric
	LatestMetrics  []LatestMetric
}

// IsLifecycleStageActive makes check that Run is in LifecycleStageActive stage.
func (r Run) IsLifecycleStageActive() bool {
	return r.LifecycleStage == LifecycleStageActive
}

type RowNum int64

func (rn *RowNum) Scan(v interface{}) error {
	nullInt := sql.NullInt64{}
	if err := nullInt.Scan(v); err != nil {
		return err
	}
	*rn = RowNum(nullInt.Int64)
	return nil
}

func (rn *RowNum) GormDataType() string {
	return "bigint"
}

func (rn *RowNum) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL: "(SELECT COALESCE(MAX(row_num), -1) FROM runs) + 1",
	}
}

type Param struct {
	Key   string `gorm:"type:varchar(250);not null;primaryKey"`
	Value string `gorm:"type:varchar(500);not null"`
	RunID string `gorm:"column:run_uuid;not null;primaryKey;index"`
}

type Tag struct {
	Key   string `gorm:"type:varchar(250);not null;primaryKey"`
	Value string `gorm:"type:varchar(5000)"`
	RunID string `gorm:"column:run_uuid;not null;primaryKey;index"`
}

type Metric struct {
	Key       string  `gorm:"type:varchar(250);not null;primaryKey"`
	Value     float64 `gorm:"type:double precision;not null;primaryKey"`
	Timestamp int64   `gorm:"not null;primaryKey"`
	RunID     string  `gorm:"column:run_uuid;not null;primaryKey;index"`
	Step      int64   `gorm:"default:0;not null;primaryKey"`
	IsNan     bool    `gorm:"default:false;not null;primaryKey"`
	Iter      int64   `gorm:"index"`
}

type LatestMetric struct {
	Key       string  `gorm:"type:varchar(250);not null;primaryKey"`
	Value     float64 `gorm:"type:double precision;not null"`
	Timestamp int64
	Step      int64  `gorm:"not null"`
	IsNan     bool   `gorm:"not null"`
	RunID     string `gorm:"column:run_uuid;not null;primaryKey;index"`
	LastIter  int64
}
