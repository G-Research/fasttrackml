package helpers

import (
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
)

var db *gorm.DB

type BaseTestSuite struct {
	suite.Suite
	AIMClient          func() *HttpClient
	MlflowClient       func() *HttpClient
	AdminClient        func() *HttpClient
	AppFixtures        *fixtures.AppFixtures
	RunFixtures        *fixtures.RunFixtures
	TagFixtures        *fixtures.TagFixtures
	ParamFixtures      *fixtures.ParamFixtures
	MetricFixtures     *fixtures.MetricFixtures
	ContextFixtures    *fixtures.ContextFixtures
	ProjectFixtures    *fixtures.ProjectFixtures
	DashboardFixtures  *fixtures.DashboardFixtures
	ExperimentFixtures *fixtures.ExperimentFixtures
	NamespaceFixtures  *fixtures.NamespaceFixtures
}

func (s *BaseTestSuite) SetupTest() {
	if db == nil {
		instance, err := database.NewDBProvider(
			GetDatabaseUri(),
			1*time.Second,
			20,
		)
		require.Nil(s.T(), err)
		db = instance.GormDB()
	}

	s.AIMClient = func() *HttpClient {
		return NewAimApiClient(GetServiceUri())
	}
	s.MlflowClient = func() *HttpClient {
		return NewMlflowApiClient(GetServiceUri())
	}
	s.AdminClient = func() *HttpClient {
		return NewAdminApiClient(GetServiceUri())
	}

	appFixtures, err := fixtures.NewAppFixtures(db)
	require.Nil(s.T(), err)
	s.AppFixtures = appFixtures

	dashboardFixtures, err := fixtures.NewDashboardFixtures(db)
	require.Nil(s.T(), err)
	s.DashboardFixtures = dashboardFixtures

	experimentFixtures, err := fixtures.NewExperimentFixtures(db)
	require.Nil(s.T(), err)
	s.ExperimentFixtures = experimentFixtures

	metricFixtures, err := fixtures.NewMetricFixtures(db)
	require.Nil(s.T(), err)
	s.MetricFixtures = metricFixtures

	namespaceFixtures, err := fixtures.NewNamespaceFixtures(db)
	require.Nil(s.T(), err)
	s.NamespaceFixtures = namespaceFixtures

	projectFixtures, err := fixtures.NewProjectFixtures(db)
	require.Nil(s.T(), err)
	s.ProjectFixtures = projectFixtures

	paramFixtures, err := fixtures.NewParamFixtures(db)
	require.Nil(s.T(), err)
	s.ParamFixtures = paramFixtures

	runFixtures, err := fixtures.NewRunFixtures(db)
	require.Nil(s.T(), err)
	s.RunFixtures = runFixtures

	tagFixtures, err := fixtures.NewTagFixtures(db)
	require.Nil(s.T(), err)
	s.TagFixtures = tagFixtures

	contextFixtures, err := fixtures.NewContextFixtures(db)
	require.Nil(s.T(), err)
	s.ContextFixtures = contextFixtures

	// by default, unload everything.
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}
