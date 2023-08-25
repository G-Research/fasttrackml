//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
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
	inputDB            *database.DbInstance
	outputDB           *database.DbInstance
	populatedRowCounts rowCounts
}

func TestImportTestSuite(t *testing.T) {
	suite.Run(t, new(ImportTestSuite))
}

func (s *ImportTestSuite) SetupTest() {
	// prepare input database.
	inputDBInstance, err := database.ConnectDB(
		helpers.GetInputDatabaseUri(),
		1*time.Second,
		20,
		false,
		false,
		"",
	)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), database.CheckAndMigrateDB(inputDBInstance, true))
	assert.Nil(s.T(), database.CreateDefaultNamespace(inputDBInstance))
	assert.Nil(s.T(), database.CreateDefaultExperiment(inputDBInstance, "s3://fasttrackml"))
	s.inputDB = inputDBInstance

	inputExperimentFixtures, err := fixtures.NewExperimentFixtures(inputDBInstance.DB)
	assert.Nil(s.T(), err)
	inputRunFixtures, err := fixtures.NewRunFixtures(inputDBInstance.DB)
	assert.Nil(s.T(), err)
	s.inputRunFixtures = inputRunFixtures

	// experiment 1
	experiment, err := inputExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    1,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	runs, err := inputRunFixtures.CreateExampleRuns(context.Background(), experiment, 5)
	assert.Nil(s.T(), err)
	s.runs = runs

	// experiment 2
	experiment, err = inputExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		NamespaceID:    1,
		LifecycleStage: models.LifecycleStageActive,
	})
	assert.Nil(s.T(), err)

	runs, err = inputRunFixtures.CreateExampleRuns(context.Background(), experiment, 5)
	assert.Nil(s.T(), err)
	s.runs = runs

	appFixtures, err := fixtures.NewAppFixtures(inputDBInstance.DB)
	app, err := appFixtures.CreateApp(context.Background(), &database.App{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		Type:  "mpi",
		State: database.AppState{},
	})

	dashboardFixtures, err := fixtures.NewDashboardFixtures(inputDBInstance.DB)
	_, err = dashboardFixtures.CreateDashboard(context.Background(), &database.Dashboard{
		Base: database.Base{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		AppID: &app.ID,
		Name:  uuid.NewString(),
	})

	// prepare output database.
	outputDBInstance, err := database.ConnectDB(
		helpers.GetOutputDatabaseUri(),
		1*time.Second,
		20,
		false,
		false,
		"",
	)
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), database.CheckAndMigrateDB(outputDBInstance, true))
	assert.Nil(s.T(), database.CreateDefaultNamespace(outputDBInstance))
	assert.Nil(s.T(), database.CreateDefaultExperiment(outputDBInstance, "s3://fasttrackml"))
	s.outputDB = outputDBInstance

	outputRunFixtures, err := fixtures.NewRunFixtures(outputDBInstance.DB)
	assert.Nil(s.T(), err)
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
	validateRowCounts(s.T(), s.inputDB, s.populatedRowCounts)

	// initially, dest DB is empty
	validateRowCounts(s.T(), s.outputDB, rowCounts{namespaces: 1, experiments: 1})

	// invoke the Importer.Import() method
	importer := database.NewImporter(s.inputDB, s.outputDB)
	err := importer.Import()
	assert.Nil(s.T(), err)

	// dest DB should now have the expected
	validateRowCounts(s.T(), s.outputDB, s.populatedRowCounts)

	// invoke the Importer.Import method a 2nd time
	err = importer.Import()
	assert.Nil(s.T(), err)

	// dest DB should still only have the expected
	validateRowCounts(s.T(), s.outputDB, s.populatedRowCounts)

	// confirm row-for-row equality
	for _, table := range []string{
		"namespaces",
		// "apps",
		// "dashboards",
		"experiment_tags",
		"runs",
		"tags",
		"params",
		"metrics",
		"latest_metrics",
	} {
		validateTable(s.T(), s.inputDB, s.outputDB, table)
	}
}

// validateRowCounts will make assertions about the db based on the test setup.
// a db imported from the test setup db should also pass these
// assertions.
func validateRowCounts(t *testing.T, db *database.DbInstance, counts rowCounts) {
	var countVal int64
	tx := db.DB.Model(&models.Namespace{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.namespaces, int(countVal), "Namespaces count incorrect")

	tx = db.DB.Model(&models.Experiment{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.experiments, int(countVal), "Experiments count incorrect")

	tx = db.DB.Model(&models.Run{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.runs, int(countVal), "Runs count incorrect")

	tx = db.DB.Model(&models.Metric{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.metrics, int(countVal), "Metrics count incorrect")

	tx = db.DB.Model(&models.LatestMetric{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.latestMetrics, int(countVal), "Latest metrics count incorrect")

	tx = db.DB.Model(&models.Tag{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.tags, int(countVal), "Run tags count incorrect")

	tx = db.DB.Model(&models.Param{}).Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.params, int(countVal), "Run params count incorrect")

	tx = db.DB.Model(&models.Run{}).Distinct("experiment_id").Count(&countVal)
	assert.Nil(t, tx.Error)
	assert.Equal(t, counts.distinctRunExperimentIDs, int(countVal), "Runs experiment association incorrect")

	// tx = db.DB.Model(&database.App{}).Count(&countVal)
	// assert.Nil(t, tx.Error)
	// assert.Equal(t, counts.apps, int(countVal), "Apps count incorrect")

	// tx = db.DB.Model(&database.Dashboard{}).Count(&countVal)
	// assert.Nil(t, tx.Error)
	// assert.Equal(t, counts.dashboards, int(countVal), "Dashboard count incorrect")
}

// validateTable will scan source and dest table and confirm they are identical
func validateTable(t *testing.T, source, dest *database.DbInstance, table string) {
	sourceRows, err := source.DB.Table(table).Rows()
	assert.Nil(t, err)
	destRows, err := dest.DB.Table(table).Rows()
	assert.Nil(t, err)
	defer sourceRows.Close()
	defer destRows.Close()

	for sourceRows.Next() {
		var sourceItem, destItem map[string]any

		err := source.DB.ScanRows(sourceRows, &sourceItem)
		assert.Nil(t, err)

		destRows.Next()
		err = dest.DB.ScanRows(destRows, &destItem)
		assert.Nil(t, err)

		//TODO:DSuhinin delete this fields right now, because they
		// cause comparison error when we compare `namespace` entities. Let's find smarter way to do that.
		delete(destItem, "updated_at")
		delete(destItem, "created_at")
		delete(sourceItem, "updated_at")
		delete(sourceItem, "created_at")

		assert.Equal(t, sourceItem, destItem)
	}
}
