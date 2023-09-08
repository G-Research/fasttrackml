package helpers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
)

var db *gorm.DB

type BaseTestSuite struct {
	AIMClient          *HttpClient
	MlflowClient       *HttpClient
	AppFixtures        *fixtures.AppFixtures
	TagFixtures        *fixtures.TagFixtures
	RunFixtures        *fixtures.RunFixtures
	ParamFixtures      *fixtures.ParamFixtures
	MetricFixtures     *fixtures.MetricFixtures
	NamespaceFixtures  *fixtures.NamespaceFixtures
	DashboardFixtures  *fixtures.DashboardFixtures
	ExperimentFixtures *fixtures.ExperimentFixtures
}

func (s *BaseTestSuite) SetupTest(t *testing.T) {
	if db == nil {
		instance, err := database.NewDBProvider(
			GetDatabaseUri(),
			1*time.Second,
			20,
			false,
		)
		assert.Nil(t, err)
		db = instance.GormDB()
	}

	appFixtures, err := fixtures.NewAppFixtures(db)
	assert.Nil(t, err)
	s.AppFixtures = appFixtures

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

	experimentFixtures, err := fixtures.NewExperimentFixtures(db)
	assert.Nil(t, err)
	s.ExperimentFixtures = experimentFixtures

	dashboardFixtures, err := fixtures.NewDashboardFixtures(db)
	assert.Nil(t, err)
	s.DashboardFixtures = dashboardFixtures

	namespaceFixtures, err := fixtures.NewNamespaceFixtures(db)
	assert.Nil(t, err)
	s.NamespaceFixtures = namespaceFixtures

	s.AIMClient = NewAimApiClient(GetServiceUri())
	s.MlflowClient = NewMlflowApiClient(GetServiceUri())

	// by default, unload everything.
	assert.Nil(t, s.NamespaceFixtures.UnloadFixtures())
}
