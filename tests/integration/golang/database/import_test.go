//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type rowCounts struct {
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
	client             *helpers.HttpClient
	inputRunFixtures   *fixtures.RunFixtures
	outputRunFixtures  *fixtures.RunFixtures
	inputDB            *database.DbInstance
	outputDB           *database.DbInstance
	populatedRowCounts rowCounts
}

func TestImportTestSuite(t *testing.T) {
	suite.Run(t, new(ImportTestSuite))
}

func (s *ImportTestSuite) SetupTest() {
	s.client = helpers.NewAimApiClient(helpers.GetServiceUri())
	inputRunFixtures, err := fixtures.NewRunFixtures(helpers.GetInputDatabaseUri())
	assert.Nil(s.T(), err)
	inputExperimentFixtures, err := fixtures.NewExperimentFixtures(helpers.GetInputDatabaseUri())
	assert.Nil(s.T(), err)
	s.inputRunFixtures = inputRunFixtures

	outputRunFixtures, err := fixtures.NewRunFixtures(helpers.GetOutputDatabaseUri())
	assert.Nil(s.T(), err)
	s.outputRunFixtures = outputRunFixtures

	// experiment 1
	experiment, err := inputExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	runs, err := inputRunFixtures.CreateExampleRuns(context.Background(), experiment, 5)
	assert.Nil(s.T(), err)
	s.runs = runs

	// experiment 2
	experiment, err = inputExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	runs, err = inputRunFixtures.CreateExampleRuns(context.Background(), experiment, 5)
	assert.Nil(s.T(), err)
	s.runs = runs

	appFixtures, err := fixtures.NewAppFixtures(helpers.GetInputDatabaseUri())
	app, err := appFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:  "mpi",
		State: database.AppState{},
	})

	dashboardFixtures, err := fixtures.NewDashboardFixtures(helpers.GetInputDatabaseUri())
	_, err = dashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		AppID: &app.ID,
		Name:  uuid.NewString(),
	})

	databaseSlowThreshold := time.Second * 1
	databasePoolMax := 20
	databaseReset := false
	databaseMigrate := false
	artifactRoot := "s3://fasttrackml"
	input, err := database.MakeDBInstance(
		helpers.GetInputDatabaseUri(),
		databaseSlowThreshold,
		databasePoolMax,
		databaseReset,
		databaseMigrate,
		artifactRoot,
	)
	assert.Nil(s.T(), err)
	output, err := database.MakeDBInstance(
		helpers.GetOutputDatabaseUri(),
		databaseSlowThreshold,
		databasePoolMax,
		databaseReset,
		databaseMigrate,
		artifactRoot,
	)
	assert.Nil(s.T(), err)

	s.inputDB = input
	s.outputDB = output

	s.populatedRowCounts = rowCounts{
		experiments:              3,
		runs:                     10,
		distinctRunExperimentIDs: 2,
		metrics:                  40,
		latestMetrics:            20,
		tags:                     10,
		params:                   20,
		dashboards:               1,
		apps:                     1,
	}
}

func (s *ImportTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.inputRunFixtures.UnloadFixtures())
		assert.Nil(s.T(), s.outputRunFixtures.UnloadFixtures())
	}()

	// source DB should have expected
	validateDB(s.T(), s.inputDB, s.populatedRowCounts)

	// initially, dest DB is empty
	validateDB(s.T(), s.outputDB, rowCounts{experiments: 1})

	// invoke the Importer.Import() method
	importer := database.NewImporter(s.inputDB, s.outputDB)
	err := importer.Import()
	assert.Nil(s.T(), err)

	// dest DB should now have the expected
	validateDB(s.T(), s.outputDB, s.populatedRowCounts)

	// invoke the Importer.Import method a 2nd time
	err = importer.Import()
	assert.Nil(s.T(), err)

	// dest DB should still only have the expected
	validateDB(s.T(), s.outputDB, s.populatedRowCounts)
}

// validateDB will make assertions about the db based on the test setup.
// a db imported from the test setup db should also pass these
// assertions.
func validateDB(t *testing.T, db *database.DbInstance, counts rowCounts) {
	var countVal int64
	tx := db.DB.Model(&database.Experiment{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.experiments, int(countVal), "Experiments count incorrect")

	tx = db.DB.Model(&database.Run{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.runs, int(countVal), "Runs count incorrect")

	tx = db.DB.Model(&database.Metric{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.metrics, int(countVal), "Metrics count incorrect")

	tx = db.DB.Model(&database.LatestMetric{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.latestMetrics, int(countVal), "Latest metrics count incorrect")

	tx = db.DB.Model(&database.Tag{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.tags, int(countVal), "Run tags count incorrect")

	tx = db.DB.Model(&database.Param{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.params, int(countVal), "Run params count incorrect")

	tx = db.DB.Model(&database.Run{}).Distinct("experiment_id").Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.distinctRunExperimentIDs, int(countVal), "Runs experiment association incorrect")

	// tx = db.DB.Model(&database.App{}).Count(&countVal)
	// assert.Nil(t, tx.Error)
	// assert.Equal(t, counts.apps, int(countVal), "Apps count incorrect")

	// tx = db.DB.Model(&database.Dashboard{}).Count(&countVal)
	// assert.Nil(t, tx.Error)
	// assert.Equal(t, counts.dashboards, int(countVal), "Dashboard count incorrect")
}
