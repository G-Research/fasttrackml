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
	AIMClient          *HttpClient
	MlflowClient       *HttpClient
	AdminClient        *HttpClient
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

	s.AIMClient = NewAimApiClient(GetServiceUri())
	s.MlflowClient = NewMlflowApiClient(GetServiceUri())
	s.AdminClient = NewAdminApiClient(GetServiceUri())

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

	// by default, unload everything.
	require.Nil(s.T(), s.NamespaceFixtures.UnloadFixtures())
}
