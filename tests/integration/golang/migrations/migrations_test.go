package migrations

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type MigrationsTestSuite struct {
	suite.Suite
}

func TestMigrationsTestSuite(t *testing.T) {
	t.Skip()
	suite.Run(t, new(MigrationsTestSuite))
}

func (s *MigrationsTestSuite) TestMigrations() {

	type Experiment struct {
		Name           string `gorm:"type:varchar(256);not null;unique"`
		ExperimentID   int32  `gorm:"column:experiment_id;not null;primaryKey"`
		LifecycleStage string `gorm:"type:varchar(32);check:lifecycle_stage IN ('active', 'deleted')"`
	}

	//nolint:lll
	type Run struct {
		ID             string `gorm:"<-:create;column:run_uuid;type:varchar(32);not null;primaryKey"`
		Name           string `gorm:"type:varchar(250)"`
		SourceType     string `gorm:"<-:create;type:varchar(20);check:source_type IN ('NOTEBOOK', 'JOB', 'LOCAL', 'UNKNOWN', 'PROJECT')"`
		Status         string `gorm:"type:varchar(9);check:status IN ('SCHEDULED', 'FAILED', 'FINISHED', 'RUNNING', 'KILLED')"`
		ExperimentID   int32
		LifecycleStage string `gorm:"type:varchar(20);check:lifecycle_stage IN ('active', 'deleted')"`
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

	tests := []struct {
		name     string
		init     func() *gorm.DB
		load     func(*gorm.DB)
		duration time.Duration
	}{
		{
			name: "TestMigrationsOverSQLite",
			init: func() *gorm.DB {
				dsn, err := helpers.GenerateDatabaseURI(s.T(), sqlite.Dialector{}.Name())
				s.Require().Nil(err)
				db, err := database.NewDBProvider(
					dsn,
					1*time.Second,
					20,
				)
				s.Require().Nil(err)

				//nolint:gosec
				mlflowSql, err := os.ReadFile("sqlite-schema.sql")
				s.Require().Nil(err)

				s.Require().Nil(db.GormDB().Exec(string(mlflowSql)).Error)
				return db.GormDB()
			},
			load: func(db *gorm.DB) {
				experiment := Experiment{
					Name:           uuid.New().String(),
					LifecycleStage: "active",
				}
				s.Require().Nil(db.Create(&experiment).Error)

				run := Run{
					Name:           uuid.New().String(),
					Status:         "RUNNING",
					ID:             uuid.New().String(),
					SourceType:     "JOB",
					LifecycleStage: "active",
					ExperimentID:   experiment.ExperimentID,
				}
				s.Require().Nil(db.Create(&run).Error)

				metrics := make([]Metric, 1_000_000)
				latestMetrics := make([]LatestMetric, 1_000_000)
				for i := 0; i < 1_000_000; i++ {
					metrics[i] = Metric{
						Key:       uuid.New().String(),
						Value:     float64(time.Now().UnixNano()),
						Timestamp: time.Now().UnixNano(),
						RunID:     run.ID,
						Step:      int64(i),
						IsNan:     false,
					}
					latestMetrics[i] = LatestMetric{
						Key:       uuid.New().String(),
						Value:     float64(time.Now().UnixNano()),
						Timestamp: time.Now().UnixNano(),
						RunID:     run.ID,
						Step:      int64(i),
						IsNan:     false,
					}
				}

				s.Require().Nil(db.CreateInBatches(metrics, 5000).Error)
				s.Require().Nil(db.CreateInBatches(latestMetrics, 5000).Error)
			},
			duration: 1*time.Minute + 10*time.Second,
		},
		{
			name: "TestMigrationsOverPostgres",
			init: func() *gorm.DB {
				dsn, err := helpers.GenerateDatabaseURI(s.T(), postgres.Dialector{}.Name())
				s.Require().Nil(err)
				db, err := database.NewDBProvider(
					dsn,
					1*time.Second,
					20,
				)
				s.Require().Nil(err)

				//nolint:gosec
				mlflowSql, err := os.ReadFile("postgres-schema.sql")
				s.Require().Nil(err)

				s.Require().Nil(db.GormDB().Exec(string(mlflowSql)).Error)
				return db.GormDB()
			},
			load: func(db *gorm.DB) {
				experiment := Experiment{
					Name:           uuid.New().String(),
					LifecycleStage: "active",
				}
				s.Require().Nil(db.Create(&experiment).Error)

				run := Run{
					ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
					Name:           uuid.New().String(),
					Status:         "RUNNING",
					SourceType:     "JOB",
					LifecycleStage: "active",
					ExperimentID:   experiment.ExperimentID,
				}
				s.Require().Nil(db.Create(&run).Error)

				metrics := make([]Metric, 1_000_000)
				latestMetrics := make([]LatestMetric, 1_000_000)
				for i := 0; i < 1_000_000; i++ {
					metrics[i] = Metric{
						Key:       uuid.New().String(),
						Value:     float64(time.Now().UnixNano()),
						Timestamp: time.Now().UnixNano(),
						RunID:     run.ID,
						Step:      int64(i),
						IsNan:     false,
					}
					latestMetrics[i] = LatestMetric{
						Key:       uuid.New().String(),
						Value:     float64(time.Now().UnixNano()),
						Timestamp: time.Now().UnixNano(),
						RunID:     run.ID,
						Step:      int64(i),
						IsNan:     false,
					}
				}

				s.Require().Nil(db.CreateInBatches(metrics, 6000).Error)
				s.Require().Nil(db.CreateInBatches(latestMetrics, 6000).Error)
			},
			duration: 50 * time.Second,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// init database schema.
			db := tt.init()

			// load test data based on current database.
			tt.load(db)

			// run migration over database and check duration time.
			start := time.Now()
			s.Require().Nil(database.CheckAndMigrateDB(true, db))
			assert.Less(s.T(), time.Since(start), tt.duration)
		})
	}
}
