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
	"github.com/G-Research/fasttrackml/pkg/database/migrations/v_0001"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type MigrationsTestSuite struct {
	suite.Suite
}

func TestMigrationsTestSuite(t *testing.T) {
	// current test is a bit slow test, so if `FML_SLOW_TESTS_ENABLED`
	// has been provided, only then run current test during the main pipeline.
	if helpers.GetSlowTestsEnabledFlag() {
		suite.Run(t, new(MigrationsTestSuite))
	}
}

func (s *MigrationsTestSuite) TestMigrations() {
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
				mlflowSql, err := os.ReadFile("mlflow-sqlite-7f2a7d5fae7d-v2.8.0.sql")
				s.Require().Nil(err)

				s.Require().Nil(db.GormDB().Exec(string(mlflowSql)).Error)
				return db.GormDB()
			},
			load: func(db *gorm.DB) {
				experiment := v_0001.Experiment{
					Name:           uuid.New().String(),
					LifecycleStage: "active",
				}
				s.Require().Nil(db.Create(&experiment).Error)

				run := v_0001.Run{
					Name:           uuid.New().String(),
					Status:         "RUNNING",
					ID:             uuid.New().String(),
					SourceType:     "JOB",
					LifecycleStage: "active",
					ExperimentID:   *experiment.ID,
				}
				s.Require().Nil(db.Omit("row_num").Create(&run).Error)

				metrics := make([]v_0001.Metric, 1_000_000)
				latestMetrics := make([]v_0001.LatestMetric, 1_000_000)
				for i := 0; i < 1_000_000; i++ {
					metrics[i] = v_0001.Metric{
						Key:       uuid.New().String(),
						Value:     float64(time.Now().UnixNano()),
						Timestamp: time.Now().UnixNano(),
						RunID:     run.ID,
						Step:      int64(i),
						IsNan:     false,
					}
					latestMetrics[i] = v_0001.LatestMetric{
						Key:       uuid.New().String(),
						Value:     float64(time.Now().UnixNano()),
						Timestamp: time.Now().UnixNano(),
						RunID:     run.ID,
						Step:      int64(i),
						IsNan:     false,
					}
				}

				s.Require().Nil(db.Omit("iter").CreateInBatches(metrics, 5000).Error)
				s.Require().Nil(db.Omit("last_iter").CreateInBatches(latestMetrics, 5000).Error)
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
				mlflowSql, err := os.ReadFile("mlflow-postgres-7f2a7d5fae7d-v2.8.0.sql")
				s.Require().Nil(err)

				s.Require().Nil(db.GormDB().Exec(string(mlflowSql)).Error)
				return db.GormDB()
			},
			load: func(db *gorm.DB) {
				experiment := v_0001.Experiment{
					Name:           uuid.New().String(),
					LifecycleStage: "active",
				}
				s.Require().Nil(db.Create(&experiment).Error)

				run := v_0001.Run{
					ID:             strings.ReplaceAll(uuid.New().String(), "-", ""),
					Name:           uuid.New().String(),
					Status:         "RUNNING",
					SourceType:     "JOB",
					LifecycleStage: "active",
					ExperimentID:   *experiment.ID,
				}
				s.Require().Nil(db.Omit("row_num").Create(&run).Error)

				metrics := make([]v_0001.Metric, 1_000_000)
				latestMetrics := make([]v_0001.LatestMetric, 1_000_000)
				for i := 0; i < 1_000_000; i++ {
					metrics[i] = v_0001.Metric{
						Key:       uuid.New().String(),
						Value:     float64(time.Now().UnixNano()),
						Timestamp: time.Now().UnixNano(),
						RunID:     run.ID,
						Step:      int64(i),
						IsNan:     false,
					}
					latestMetrics[i] = v_0001.LatestMetric{
						Key:       uuid.New().String(),
						Value:     float64(time.Now().UnixNano()),
						Timestamp: time.Now().UnixNano(),
						RunID:     run.ID,
						Step:      int64(i),
						IsNan:     false,
					}
				}

				s.Require().Nil(db.Omit("iter").CreateInBatches(metrics, 5000).Error)
				s.Require().Nil(db.Omit("last_iter").CreateInBatches(latestMetrics, 5000).Error)
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
