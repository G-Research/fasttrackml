//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type ImportTestSuite struct {
	suite.Suite
	runs              []*models.Run
	client            *helpers.HttpClient
	inputRunFixtures  *fixtures.RunFixtures
	outputRunFixtures *fixtures.RunFixtures
	inputDB           *database.DbInstance
	outputDB          *database.DbInstance
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

	experiment, err := inputExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	runs, err := inputRunFixtures.CreateExampleRuns(context.Background(), experiment, 5)
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

}

func (s *ImportTestSuite) Test_Ok() {
	defer func() {
		assert.Nil(s.T(), s.inputRunFixtures.UnloadFixtures())
		assert.Nil(s.T(), s.outputRunFixtures.UnloadFixtures())
	}()

	validateDB(s.T(), s.inputDB)

	// initially, ouput db has 0 rows
	outputRuns, err := s.outputRunFixtures.GetRuns(context.Background(), s.runs[0].ExperimentID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(outputRuns))

	// invoke the Import method
	database.Import(s.inputDB, s.outputDB)

	validateDB(s.T(), s.outputDB)
}

// validateDB will make assertions about the db based on the test setup.
// a db imported from the test setup db should also pass these
// assertions.
func validateDB(t *testing.T, db *database.DbInstance) {

	numberOfRuns := 5
	numberOfMetrics := 20
	numberOfLatestMetrics := 10
	numberOfTags := 5
	numberOfParams := 0
	
	var countVal int64
	tx := db.DB.Model(&database.Run{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, numberOfRuns, int(countVal))

	tx = db.DB.Model(&database.Metric{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, numberOfMetrics, int(countVal))

	tx = db.DB.Model(&database.LatestMetric{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, numberOfLatestMetrics, int(countVal))

	tx = db.DB.Model(&database.Tag{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, numberOfTags, int(countVal))

	tx = db.DB.Model(&database.Param{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, numberOfParams, int(countVal))
}
