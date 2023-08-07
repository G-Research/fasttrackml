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

const (
	numberOfRuns = 5
)

type ImportTestSuite struct {
	suite.Suite
	runs                     []*models.Run
	client                   *helpers.HttpClient
	inputRunFixtures  *fixtures.RunFixtures
	outputRunFixtures *fixtures.RunFixtures
	inputDB           *database.DbInstance
	outputDB           *database.DbInstance
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

	runs, err := inputRunFixtures.CreateExampleRuns(context.Background(), experiment, numberOfRuns)
	assert.Nil(s.T(), err)
	s.runs = runs

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

	runs, err := s.inputRunFixtures.GetTestRuns(context.Background(), s.runs[0].ExperimentID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), numberOfRuns, len(runs))

	runs, err = s.outputRunFixtures.GetTestRuns(context.Background(), s.runs[0].ExperimentID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(runs))

	// invoke the subject method
	database.Import(s.inputDB, s.outputDB, false)

	runs, err = s.outputRunFixtures.GetTestRuns(context.Background(), s.runs[0].ExperimentID)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), numberOfRuns, len(runs))
}
