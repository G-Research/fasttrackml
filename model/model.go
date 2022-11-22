package model

import (
	"database/sql"
	"encoding/hex"

	"github.com/google/uuid"
)

type Status string

const (
	StatusRunning   Status = "RUNNING"
	StatusScheduled Status = "SCHEDULED"
	StatusFinished  Status = "FINISHED"
	StatusFailed    Status = "FAILED"
	StatusKilled    Status = "KILLED"
)

type LifecycleStage string

const (
	LifecycleStageActive  LifecycleStage = "active"
	LifecycleStageDeleted LifecycleStage = "deleted"
)

type Experiment struct {
	ID               *int32         `gorm:"column:experiment_id;not null;primaryKey"`
	Name             string         `gorm:"type:varchar(256);not null;unique"`
	ArtifactLocation string         `gorm:"type:varchar(256)"`
	LifecycleStage   LifecycleStage `gorm:"type:varchar(32);check:lifecycle_stage IN ('active', 'deleted')"`
	CreationTime     sql.NullInt64  `gorm:"type:bigint"`
	LastUpdateTime   sql.NullInt64  `gorm:"type:bigint"`
	Tags             []ExperimentTag
	Runs             []Run
}

type ExperimentTag struct {
	Key          string `gorm:"type:varchar(250);not null;primaryKey"`
	Value        string `gorm:"type:varchar(5000)"`
	ExperimentID int32  `gorm:"not null;primaryKey"`
}

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
	DeletedTime    sql.NullInt64 `gorm:"type:bigint"`
	Params         []Param
	Tags           []Tag
	Metrics        []Metric
	LatestMetrics  []LatestMetric
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
}

type LatestMetric struct {
	Key       string  `gorm:"type:varchar(250);not null;primaryKey"`
	Value     float64 `gorm:"type:double precision;not null"`
	Timestamp int64
	Step      int64  `gorm:"not null"`
	IsNan     bool   `gorm:"not null"`
	RunID     string `gorm:"column:run_uuid;not null;primaryKey;index"`
}

type AlembicVersion struct {
	Version string `gorm:"column:version_num;type:varchar(32);not null;primaryKey"`
}

func (AlembicVersion) TableName() string {
	return "alembic_version"
}

func NewUUID() string {
	var r [32]byte
	u := uuid.New()
	hex.Encode(r[:], u[:])
	return string(r[:])
}
