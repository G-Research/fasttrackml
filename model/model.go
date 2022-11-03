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
	ID               *int32         `gorm:"column:experiment_id;primaryKey"`
	Name             string         `gorm:"type:varchar(256);not null;uniqueIndex"`
	ArtifactLocation string         `gorm:"type:varchar(256)"`
	LifecycleStage   LifecycleStage `gorm:"type:varchar(32)"`
	CreationTime     sql.NullInt64
	LastUpdateTime   sql.NullInt64
	Tags             []ExperimentTag
	Runs             []Run
}

type ExperimentTag struct {
	Key          string `gorm:"type:varchar(250);primaryKey"`
	Value        string `gorm:"type:varchar(5000)"`
	ExperimentID int32  `gorm:"primaryKey"`
}

type Run struct {
	ID             string `gorm:"column:run_uuid;type:varchar(32);primaryKey"`
	Name           string `gorm:"type:varchar(250)"`
	SourceType     string `gorm:"type:varchar(20);default:UNKNOWN"`
	SourceName     string `gorm:"type:varchar(500)"`
	EntryPointName string `gorm:"type:varchar(50)"`
	UserID         string `gorm:"type:varchar(256)"`
	Status         Status `gorm:"type:varchar(9)"`
	StartTime      sql.NullInt64
	EndTime        sql.NullInt64
	SourceVersion  string         `gorm:"type:varchar(50)"`
	LifecycleStage LifecycleStage `gorm:"type:varchar(20)"`
	ArtifactURI    string         `gorm:"type:varchar(200)"`
	ExperimentID   int32
	DeletedTime    sql.NullInt64
	Params         []Param
	Tags           []Tag
	Metrics        []Metric
	LatestMetrics  []LatestMetric
}

type Param struct {
	Key   string `gorm:"type:varchar(250);primaryKey"`
	Value string `gorm:"type:varchar(500);not null"`
	RunID string `gorm:"column:run_uuid;primaryKey;index"`
}

type Tag struct {
	Key   string `gorm:"type:varchar(250);primaryKey"`
	Value string `gorm:"type:varchar(5000)"`
	RunID string `gorm:"column:run_uuid;primaryKey;index"`
}

type Metric struct {
	Key       string  `gorm:"type:varchar(250);primaryKey"`
	Value     float64 `gorm:"type:double precision;primaryKey"`
	Timestamp int64   `gorm:"primaryKey"`
	RunID     string  `gorm:"column:run_uuid;primaryKey;index"`
	Step      int64   `gorm:"default:0;primaryKey"`
	IsNan     bool    `gorm:"default:false;primaryKey"`
}

type LatestMetric struct {
	Key       string  `gorm:"type:varchar(250);primaryKey"`
	Value     float64 `gorm:"type:double precision;not null"`
	Timestamp int64
	Step      int64  `gorm:"not null"`
	IsNan     bool   `gorm:"not null"`
	RunID     string `gorm:"column:run_uuid;primaryKey;index"`
}

func NewUUID() string {
	var r [32]byte
	u := uuid.New()
	hex.Encode(r[:], u[:])
	return string(r[:])
}
