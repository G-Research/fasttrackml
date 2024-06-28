package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/types"
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

// Default Experiment properties.
const (
	DefaultExperimentID   = int32(0)
	DefaultExperimentName = "Default"
)

type Namespace struct {
	ID                  uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Apps                []App          `gorm:"constraint:OnDelete:CASCADE" json:"apps"`
	Code                string         `gorm:"unique;index;not null" json:"code"`
	Description         string         `json:"description"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	DefaultExperimentID *int32         `gorm:"not null" json:"default_experiment_id"`
	Experiments         []Experiment   `gorm:"constraint:OnDelete:CASCADE" json:"experiments"`
}

type Experiment struct {
	ID               *int32         `gorm:"column:experiment_id;not null;primaryKey"`
	Name             string         `gorm:"type:varchar(256);not null;index:,unique,composite:name"`
	ArtifactLocation string         `gorm:"type:varchar(256)"`
	LifecycleStage   LifecycleStage `gorm:"type:varchar(32);check:lifecycle_stage IN ('active', 'deleted')"`
	CreationTime     sql.NullInt64  `gorm:"type:bigint"`
	LastUpdateTime   sql.NullInt64  `gorm:"type:bigint"`
	NamespaceID      uint           `gorm:"not null;index:,unique,composite:name"`
	Namespace        Namespace
	Tags             []ExperimentTag `gorm:"constraint:OnDelete:CASCADE"`
	Runs             []Run           `gorm:"constraint:OnDelete:CASCADE"`
}

// IsDefault makes check that Experiment is default.
func (e Experiment) IsDefault(namespace *models.Namespace) bool {
	return e.ID != nil && namespace.DefaultExperimentID != nil && *e.ID == *namespace.DefaultExperimentID
}

type ExperimentTag struct {
	Key          string `gorm:"type:varchar(250);not null;primaryKey"`
	Value        string `gorm:"type:varchar(5000)"`
	ExperimentID int32  `gorm:"not null;primaryKey"`
}

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
	Metrics        []Metric       `gorm:"constraint:OnDelete:CASCADE"`
	LatestMetrics  []LatestMetric `gorm:"constraint:OnDelete:CASCADE"`
	Logs           []Log          `gorm:"constraing:OnDelete:CASCADE"`
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

func (rn RowNum) GormDataType() string {
	return "bigint"
}

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

type Param struct {
	Key        string   `gorm:"type:varchar(250);not null;primaryKey"`
	ValueStr   *string  `gorm:"type:varchar(500)"`
	ValueInt   *int64   `gorm:"type:bigint"`
	ValueFloat *float64 `gorm:"type:float"`
	RunID      string   `gorm:"column:run_uuid;not null;primaryKey;index"`
}

// Tag represents metadata about a particular run (for Mlflow).
type Tag struct {
	Key   string `gorm:"type:varchar(250);not null;primaryKey"`
	Value string `gorm:"type:varchar(5000)"`
	RunID string `gorm:"column:run_uuid;not null;primaryKey;index"`
}

// SharedTag represents a tag which can label multiple runs (for Aim).
type SharedTag struct {
	ID          uuid.UUID `gorm:"column:id;not null;primaryKey"`
	IsArchived  bool      `gorm:"not null,default:false"`
	Name        string    `gorm:"type:varchar(250);not null"`
	Color       string    `gorm:"type:varchar(7);null"`
	Description string    `gorm:"type:varchar(500);null"`
	NamespaceID uint      `gorm:"not null"`
	Runs        []Run     `gorm:"many2many:run_shared_tags"`
}

type Metric struct {
	Key       string  `gorm:"type:varchar(250);not null;primaryKey"`
	Value     float64 `gorm:"type:double precision;not null;primaryKey"`
	Timestamp int64   `gorm:"not null;primaryKey"`
	RunID     string  `gorm:"column:run_uuid;not null;primaryKey;index"`
	Step      int64   `gorm:"default:0;not null;primaryKey"`
	IsNan     bool    `gorm:"default:false;not null;primaryKey"`
	Iter      int64   `gorm:"index"`
	ContextID uint    `gorm:"not null;primaryKey"`
	Context   Context
}

type LatestMetric struct {
	Key       string  `gorm:"type:varchar(250);not null;primaryKey"`
	Value     float64 `gorm:"type:double precision;not null"`
	Timestamp int64
	Step      int64  `gorm:"not null"`
	IsNan     bool   `gorm:"not null"`
	RunID     string `gorm:"column:run_uuid;not null;primaryKey;index"`
	LastIter  int64
	ContextID uint `gorm:"not null;primaryKey"`
	Context   Context
}

type Log struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Value     string `gorm:"not null"`
	RunID     string `gorm:"column:run_uuid;not null;index"`
	Timestamp int64  `gorm:"not null;index"`
}

type Context struct {
	ID   uint        `gorm:"primaryKey;autoIncrement"`
	Json types.JSONB `gorm:"not null;unique;index"`
}

// GetJsonHash returns hash of the Context.Json
func (c Context) GetJsonHash() string {
	hash := sha256.Sum256(c.Json)
	return string(hash[:])
}

type AlembicVersion struct {
	Version string `gorm:"column:version_num;type:varchar(32);not null;primaryKey"`
}

func (AlembicVersion) TableName() string {
	return "alembic_version"
}

type SchemaVersion struct {
	Version string `gorm:"not null;primaryKey"`
}

func (SchemaVersion) TableName() string {
	return "schema_version"
}

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	b.ID = uuid.New()
	return nil
}

type Dashboard struct {
	Base
	Name        string     `json:"name"`
	Description string     `json:"description"`
	AppID       *uuid.UUID `gorm:"type:uuid" json:"app_id"`
	App         App        `json:"-"`
	IsArchived  bool       `json:"-"`
}

func (d Dashboard) MarshalJSON() ([]byte, error) {
	type localDashboard Dashboard
	type jsonDashboard struct {
		localDashboard
		AppType *string `json:"app_type"`
	}
	jd := jsonDashboard{
		localDashboard: localDashboard(d),
	}
	if d.App.IsArchived {
		jd.AppID = nil
	} else {
		jd.AppType = &d.App.Type
	}
	return json.Marshal(jd)
}

type App struct {
	Base
	Type        string    `gorm:"not null" json:"type"`
	State       AppState  `json:"state"`
	Namespace   Namespace `json:"-"`
	NamespaceID uint      `gorm:"not null" json:"-"`
	IsArchived  bool      `json:"-"`
}

type AppState map[string]any

func (s AppState) Value() (driver.Value, error) {
	v, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

func (s *AppState) Scan(v interface{}) error {
	var nullS sql.NullString
	if err := nullS.Scan(v); err != nil {
		return err
	}
	if nullS.Valid {
		return json.Unmarshal([]byte(nullS.String), s)
	}
	return nil
}

func (s AppState) GormDataType() string {
	return "text"
}

func NewUUID() string {
	var r [32]byte
	u := uuid.New()
	hex.Encode(r[:], u[:])
	return string(r[:])
}

type Role struct {
	Base
	Name string `gorm:"unique;index;not null"`
}

type RoleNamespace struct {
	Base
	Role        Role      `gorm:"constraint:OnDelete:CASCADE"`
	RoleID      uuid.UUID `gorm:"not null;index:,unique,composite:relation"`
	Namespace   Namespace `gorm:"constraint:OnDelete:CASCADE"`
	NamespaceID uint      `gorm:"not null;index:,unique,composite:relation"`
}

type Artifact struct {
	Base
	Iter    int64 `gorm:"index"`
	Step    int64 `gorm:"default:0;not null"`
	Run     Run
	RunID   string `gorm:"column:run_uuid;not null;index;constraint:OnDelete:CASCADE"`
	Index   int64
	Width   int64
	Height  int64
	Format  string
	Caption string
	BlobURI string
}
