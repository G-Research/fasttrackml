package helpers

import (
	"context"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/tests/integration/golang/fixtures"
)

var db *gorm.DB

type BaseTestSuite struct {
	suite.Suite
	AIMClient                   func() *HttpClient
	MlflowClient                func() *HttpClient
	AdminClient                 func() *HttpClient
	AppFixtures                 *fixtures.AppFixtures
	RunFixtures                 *fixtures.RunFixtures
	TagFixtures                 *fixtures.TagFixtures
	MetricFixtures              *fixtures.MetricFixtures
	ParamFixtures               *fixtures.ParamFixtures
	ProjectFixtures             *fixtures.ProjectFixtures
	DashboardFixtures           *fixtures.DashboardFixtures
	ExperimentFixtures          *fixtures.ExperimentFixtures
	DefaultExperiment           *models.Experiment
	NamespaceFixtures           *fixtures.NamespaceFixtures
	DefaultNamespace            *models.Namespace
	ResetOnSubTest              bool
	SkipCreateDefaultNamespace  bool
	SkipCreateDefaultExperiment bool
	setupHooks                  []func()
	tearDownHooks               []func()
}

func (s *BaseTestSuite) SetupSuite() {
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

	s.AddSetupHook(s.setup)
	s.AddTearDownHook(s.tearDown)
}

func (s *BaseTestSuite) setup() {
	s.Require().Nil(s.NamespaceFixtures.TruncateTables())

	if !s.SkipCreateDefaultNamespace {
		var err error
		s.DefaultNamespace, err = s.NamespaceFixtures.CreateNamespace(context.Background(), &models.Namespace{
			Code:                "default",
			DefaultExperimentID: common.GetPointer(int32(0)),
		})
		s.Require().Nil(err)

		if !s.SkipCreateDefaultExperiment {
			s.DefaultExperiment, err = s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
				Name:           "Default",
				LifecycleStage: models.LifecycleStageActive,
				NamespaceID:    s.DefaultNamespace.ID,
			})
			s.Require().Nil(err)

			s.DefaultNamespace.DefaultExperimentID = s.DefaultExperiment.ID
			_, err = s.NamespaceFixtures.UpdateNamespace(context.Background(), s.DefaultNamespace)
			s.Require().Nil(err)
		}
	}
}

func (s *BaseTestSuite) AddSetupHook(hook func()) {
	s.setupHooks = append(s.setupHooks, hook)
}

func (s *BaseTestSuite) runSetupHooks() {
	for _, hook := range s.setupHooks {
		hook()
	}
}

func (s *BaseTestSuite) SetupTest() {
	if !s.ResetOnSubTest {
		s.runSetupHooks()
	}
}

func (s *BaseTestSuite) SetupSubTest() {
	if s.ResetOnSubTest {
		s.runSetupHooks()
	}
}

func (s *BaseTestSuite) AddTearDownHook(hook func()) {
	s.tearDownHooks = append([]func(){hook}, s.tearDownHooks...)
}

func (s *BaseTestSuite) tearDown() {
	s.Require().Nil(s.NamespaceFixtures.TruncateTables())
}

func (s *BaseTestSuite) runTearDownHooks() {
	for _, hook := range s.tearDownHooks {
		hook()
	}
}

func (s *BaseTestSuite) TearDownTest() {
	if !s.ResetOnSubTest {
		s.runTearDownHooks()
	}
}

func (s *BaseTestSuite) TearDownSubTest() {
	if s.ResetOnSubTest {
		s.runTearDownHooks()
	}
}
