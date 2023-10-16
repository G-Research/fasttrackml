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
	DashboardFixtures  *fixtures.DashboardFixtures
	ExperimentFixtures *fixtures.ExperimentFixtures
	MetricFixtures     *fixtures.MetricFixtures
	NamespaceFixtures  *fixtures.NamespaceFixtures
	ParamFixtures      *fixtures.ParamFixtures
	ProjectFixtures    *fixtures.ProjectFixtures
	RunFixtures        *fixtures.RunFixtures
	TagFixtures        *fixtures.TagFixtures
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

	s.AIMClient = NewAimApiClient(GetServiceUri())
	s.MlflowClient = NewMlflowApiClient(GetServiceUri())

	appFixtures, err := fixtures.NewAppFixtures(db)
	assert.Nil(t, err)
	s.AppFixtures = appFixtures

	dashboardFixtures, err := fixtures.NewDashboardFixtures(db)
	assert.Nil(t, err)
	s.DashboardFixtures = dashboardFixtures

	experimentFixtures, err := fixtures.NewExperimentFixtures(db)
	assert.Nil(t, err)
	s.ExperimentFixtures = experimentFixtures

	metricFixtures, err := fixtures.NewMetricFixtures(db)
	assert.Nil(t, err)
	s.MetricFixtures = metricFixtures

	namespaceFixtures, err := fixtures.NewNamespaceFixtures(db)
	assert.Nil(t, err)
	s.NamespaceFixtures = namespaceFixtures

	projectFixtures, err := fixtures.NewProjectFixtures(db)
	assert.Nil(t, err)
	s.ProjectFixtures = projectFixtures

	paramFixtures, err := fixtures.NewParamFixtures(db)
	assert.Nil(t, err)
	s.ParamFixtures = paramFixtures

	runFixtures, err := fixtures.NewRunFixtures(db)
	assert.Nil(t, err)
	s.RunFixtures = runFixtures

	tagFixtures, err := fixtures.NewTagFixtures(db)
	assert.Nil(t, err)
	s.TagFixtures = tagFixtures

	// by default, unload everything.
	assert.Nil(t, s.NamespaceFixtures.UnloadFixtures())
}
