//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type rowCounts struct {
	namespaces               int
	experiments              int
	runs                     int
	distinctRunExperimentIDs int
	metrics                  int
	latestMetrics            int
	tags                     int
	params                   int
	dashboards               int
	apps                     int
}

type ImportTestSuite struct {
	suite.Suite
	runs               []*models.Run
	inputRunFixtures   *fixtures.RunFixtures
	outputRunFixtures  *fixtures.RunFixtures
	inputDB            *gorm.DB
	outputDB           *gorm.DB
	populatedRowCounts rowCounts
}

func TestImportTestSuite(t *testing.T) {
	suite.Run(t, new(ImportTestSuite))
}

func (s *ImportTestSuite) SetupTest() {
	// prepare input database.
	db, err := database.NewDBProvider(
		helpers.GetInputDatabaseUri(),
		1*time.Second,
		20,
	)
	s.Require().Nil(err)
	s.Require().Nil(database.CheckAndMigrateDB(true, db.GormDB()))
	s.Require().Nil(database.CreateDefaultNamespace(db.GormDB()))
	s.Require().Nil(database.CreateDefaultExperiment(db.GormDB(), "s3://fasttrackml"))
	s.inputDB = db.GormDB()

	inputRunFixtures, err := fixtures.NewRunFixtures(db.GormDB())
	s.Require().Nil(err)
	s.inputRunFixtures = inputRunFixtures
	s.populateDB(s.inputDB)

	// prepare output database.
	db, err = database.NewDBProvider(
		helpers.GetOutputDatabaseUri(),
		1*time.Second,
		20,
	)
	s.Require().Nil(err)
	s.Require().Nil(database.CheckAndMigrateDB(true, db.GormDB()))
	s.Require().Nil(database.CreateDefaultNamespace(db.GormDB()))
	s.Require().Nil(database.CreateDefaultExperiment(db.GormDB(), "s3://fasttrackml"))
	s.outputDB = db.GormDB()

	outputRunFixtures, err := fixtures.NewRunFixtures(db.GormDB())
	s.Require().Nil(err)
	s.outputRunFixtures = outputRunFixtures

	s.populatedRowCounts = rowCounts{
		namespaces:               1,
		experiments:              3,
		runs:                     10,
		distinctRunExperimentIDs: 2,
		metrics:                  40,
		latestMetrics:            20,
		tags:                     10,
		params:                   20,
		dashboards:               2,
		apps:                     1,
	}
}

func (s *ImportTestSuite) populateDB(db *gorm.DB) {
	experimentFixtures, err := fixtures.NewExperimentFixtures(db)
	s.Require().Nil(err)

	runFixtures, err := fixtures.NewRunFixtures(db)
	s.Require().Nil(err)

	// experiment 1
	experiment, err := experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    1,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	runs, err := runFixtures.CreateExampleRuns(context.Background(), experiment, 5)
	s.Require().Nil(err)
	s.runs = runs

	// experiment 2
	experiment, err = experimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    1,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	runs, err = runFixtures.CreateExampleRuns(context.Background(), experiment, 5)
	s.Require().Nil(err)
	s.runs = runs

	appFixtures, err := fixtures.NewAppFixtures(db)
	s.Require().Nil(err)
	app, err := appFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		NamespaceID: 1,
		Type:        "mpi",
		State:       database.AppState{},
	})
	s.Require().Nil(err)

	dashboardFixtures, err := fixtures.NewDashboardFixtures(db)
	s.Require().Nil(err)

	// dashboard 1
	_, err = dashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		AppID: &app.ID,
		Name:  uuid.NewString(),
	})
	s.Require().Nil(err)

	// dashboard 2
	_, err = dashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		AppID: &app.ID,
		Name:  uuid.NewString(),
	})
	s.Require().Nil(err)
}

func (s *ImportTestSuite) Test_Ok() {
	defer func() {
		s.Require().Nil(s.inputRunFixtures.UnloadFixtures())
		s.Require().Nil(s.outputRunFixtures.UnloadFixtures())
	}()

	// source DB should have expected
	s.validateRowCounts(s.inputDB, s.populatedRowCounts)

	// initially, dest DB is empty
	s.validateRowCounts(s.outputDB, rowCounts{namespaces: 1, experiments: 1})

	// invoke the Importer.Import() method
	importer := database.NewImporter(s.inputDB, s.outputDB)
	s.Require().Nil(importer.Import())

	// dest DB should now have the expected
	s.validateRowCounts(s.outputDB, s.populatedRowCounts)

	// invoke the Importer.Import method a 2nd time
	s.Require().Nil(importer.Import())

	// dest DB should still only have the expected (idempotent)
	s.validateRowCounts(s.outputDB, s.populatedRowCounts)

	// confirm row-for-row equality
	for _, table := range []string{
		"namespaces",
		"apps",
		"dashboards",
		"experiment_tags",
		"runs",
		"tags",
		"params",
		"metrics",
		"latest_metrics",
	} {
		s.validateTable(s.inputDB, s.outputDB, table)
	}
}

// validateRowCounts will make assertions about the db based on the test setup.
// a db imported from the test setup db should also pass these
// assertions.
func (s *ImportTestSuite) validateRowCounts(db *gorm.DB, counts rowCounts) {
	var countVal int64
	s.Require().Nil(db.Model(&models.Namespace{}).Count(&countVal).Error)
	s.Equal(counts.namespaces, int(countVal), "Namespaces count incorrect")

	s.Require().Nil(db.Model(&models.Experiment{}).Count(&countVal).Error)
	s.Equal(counts.experiments, int(countVal), "Experiments count incorrect")

	s.Require().Nil(db.Model(&models.Run{}).Count(&countVal).Error)
	s.Equal(counts.runs, int(countVal), "Runs count incorrect")

	s.Require().Nil(db.Model(&models.Metric{}).Count(&countVal).Error)
	s.Equal(counts.metrics, int(countVal), "Metrics count incorrect")

	s.Require().Nil(db.Model(&models.LatestMetric{}).Count(&countVal).Error)
	s.Equal(counts.latestMetrics, int(countVal), "Latest metrics count incorrect")

	s.Require().Nil(db.Model(&models.Tag{}).Count(&countVal).Error)
	s.Equal(counts.tags, int(countVal), "Run tags count incorrect")

	s.Require().Nil(db.Model(&models.Param{}).Count(&countVal).Error)
	s.Equal(counts.params, int(countVal), "Run params count incorrect")

	s.Require().Nil(db.Model(&models.Run{}).Distinct("experiment_id").Count(&countVal).Error)
	s.Equal(counts.distinctRunExperimentIDs, int(countVal), "Runs experiment association incorrect")

	s.Require().Nil(db.Model(&database.App{}).Count(&countVal).Error)
	s.Equal(counts.apps, int(countVal), "Apps count incorrect")

	s.Require().Nil(db.Model(&database.Dashboard{}).Count(&countVal).Error)
	s.Equal(counts.dashboards, int(countVal), "Dashboard count incorrect")
}

// validateTable will scan source and dest table and confirm they are identical
func (s *ImportTestSuite) validateTable(source, dest *gorm.DB, table string) {
	sourceRows, err := source.Table(table).Rows()
	s.Require().Nil(err)
	s.Require().Nil(sourceRows.Err())
	destRows, err := dest.Table(table).Rows()
	s.Require().Nil(err)
	s.Require().Nil(destRows.Err())
	//nolint:errcheck
	defer sourceRows.Close()
	//nolint:errcheck
	defer destRows.Close()

	for sourceRows.Next() {
		// dest should have the same number of rows as source
		s.Require().True(destRows.Next())

		var sourceRow, destRow map[string]any
		s.Require().Nil(source.ScanRows(sourceRows, &sourceRow))
		s.Require().Nil(dest.ScanRows(destRows, &destRow))

		// TODO:DSuhinin delete this fields right now, because they
		// cause comparison error when we compare `namespace` entities. Let's find smarter way to do that.
		delete(destRow, "updated_at")
		delete(destRow, "created_at")
		delete(sourceRow, "updated_at")
		delete(sourceRow, "created_at")

		s.Equal(sourceRow, destRow)
	}
	// dest should have the same number of rows as source
	s.Require().False(destRows.Next())
}
