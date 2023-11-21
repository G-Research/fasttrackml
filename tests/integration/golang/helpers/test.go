package helpers

import (
	"time"

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
	MetricFixtures     *fixtures.MetricFixtures
	ParamFixtures      *fixtures.ParamFixtures
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
		s.Require().Nil(err)
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
	s.Require().Nil(err)
	s.AppFixtures = appFixtures

	dashboardFixtures, err := fixtures.NewDashboardFixtures(db)
	s.Require().Nil(err)
	s.DashboardFixtures = dashboardFixtures

	experimentFixtures, err := fixtures.NewExperimentFixtures(db)
	s.Require().Nil(err)
	s.ExperimentFixtures = experimentFixtures

	metricFixtures, err := fixtures.NewMetricFixtures(db)
	s.Require().Nil(err)
	s.MetricFixtures = metricFixtures

	namespaceFixtures, err := fixtures.NewNamespaceFixtures(db)
	s.Require().Nil(err)
	s.NamespaceFixtures = namespaceFixtures

	projectFixtures, err := fixtures.NewProjectFixtures(db)
	s.Require().Nil(err)
	s.ProjectFixtures = projectFixtures

	paramFixtures, err := fixtures.NewParamFixtures(db)
	s.Require().Nil(err)
	s.ParamFixtures = paramFixtures

	runFixtures, err := fixtures.NewRunFixtures(db)
	s.Require().Nil(err)
	s.RunFixtures = runFixtures

	tagFixtures, err := fixtures.NewTagFixtures(db)
	s.Require().Nil(err)
	s.TagFixtures = tagFixtures

	// by default, unload everything.
	s.Require().Nil(s.NamespaceFixtures.UnloadFixtures())
}
