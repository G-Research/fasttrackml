package helpers

import (
	"testing"
	"time"

	"github.com/zeebo/assert"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
)

var db *gorm.DB

type BaseTestSuite struct {
	MlflowClient       *HttpClient
	TagFixtures        *fixtures.TagFixtures
	RunFixtures        *fixtures.RunFixtures
	ParamFixtures      *fixtures.ParamFixtures
	MetricFixtures     *fixtures.MetricFixtures
	NamespaceFixtures  *fixtures.NamespaceFixtures
	ExperimentFixtures *fixtures.ExperimentFixtures
}

func (s *BaseTestSuite) SetupTest(t *testing.T) {
	s.MlflowClient = NewMlflowApiClient(GetServiceUri())

	if db == nil {
		instance, err := database.ConnectDB(
			GetDatabaseUri(),
			1*time.Second,
			20,
			false,
			false,
			"",
		)
		db = instance.DB
		assert.Nil(t, err)
	}

	tagFixtures, err := fixtures.NewTagFixtures(db)
	assert.Nil(t, err)
	s.TagFixtures = tagFixtures

	runFixtures, err := fixtures.NewRunFixtures(db)
	assert.Nil(t, err)
	s.RunFixtures = runFixtures

	paramFixtures, err := fixtures.NewParamFixtures(db)
	assert.Nil(t, err)
	s.ParamFixtures = paramFixtures

	metricFixtures, err := fixtures.NewMetricFixtures(db)
	assert.Nil(t, err)
	s.MetricFixtures = metricFixtures

	expFixtures, err := fixtures.NewExperimentFixtures(db)
	assert.Nil(t, err)
	s.ExperimentFixtures = expFixtures

	namespaceFixtures, err := fixtures.NewNamespaceFixtures(db)
	assert.Nil(t, err)
	s.NamespaceFixtures = namespaceFixtures

	// by default, unload everything.
	assert.Nil(t, s.NamespaceFixtures.UnloadFixtures())
}
