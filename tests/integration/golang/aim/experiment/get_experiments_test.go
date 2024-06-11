package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentsTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentsTestSuite(t *testing.T) {
	suite.Run(t, &GetExperimentsTestSuite{
		helpers.BaseTestSuite{
			SkipCreateDefaultExperiment: true,
		},
	})
}

func (s *GetExperimentsTestSuite) Test_Ok() {
	experiments := map[string]*models.Experiment{}
	for i := 0; i < 5; i++ {
		experiment := &models.Experiment{
			Name:        fmt.Sprintf("Test Experiment %d", i),
			NamespaceID: s.DefaultNamespace.ID,
			CreationTime: sql.NullInt64{
				Int64: time.Now().UTC().UnixMilli(),
				Valid: true,
			},
			LifecycleStage: models.LifecycleStageActive,
		}
		experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), experiment)
		s.Require().Nil(err)
		experiments[fmt.Sprintf("%d", *experiment.ID)] = experiment
	}

	var resp []response.Experiment
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/experiments/"))
	s.Require().Equal(len(experiments), len(resp))
	for _, actualExperiment := range resp {
		expectedExperiment := experiments[actualExperiment.ID]
		s.Equal(fmt.Sprintf("%d", *expectedExperiment.ID), actualExperiment.ID)
		s.Equal(expectedExperiment.Name, actualExperiment.Name)
		s.Equal(float64(expectedExperiment.CreationTime.Int64)/1000, actualExperiment.CreationTime)
		s.Equal(expectedExperiment.LifecycleStage == models.LifecycleStageDeleted, actualExperiment.Archived)
		s.Equal(len(expectedExperiment.Runs), actualExperiment.RunCount)
		s.Equal(helpers.GetDescriptionFromExperiment(*expectedExperiment), actualExperiment.Description)
	}
}
