package experiment

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type GetExperimentTestSuite struct {
	helpers.BaseTestSuite
}

func TestGetExperimentTestSuite(t *testing.T) {
	suite.Run(t, &GetExperimentTestSuite{
		helpers.BaseTestSuite{
			SkipCreateDefaultExperiment: true,
		},
	})
}

func (s *GetExperimentTestSuite) Test_Ok() {
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name: "Test Experiment",
		Tags: []models.ExperimentTag{
			{
				Key:   common.DescriptionTagKey,
				Value: "value1",
			},
		},
		CreationTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		NamespaceID:    s.DefaultNamespace.ID,
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)

	var resp response.GetExperiment
	s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/experiments/%d", *experiment.ID))
	s.Equal(fmt.Sprintf("%d", *experiment.ID), resp.ID)
	s.Equal(experiment.Name, resp.Name)
	s.Equal(helpers.GetDescriptionFromExperiment(*experiment), resp.Description)
	s.Equal(float64(experiment.CreationTime.Int64)/1000, resp.CreationTime)
	s.Equal(false, resp.Archived)
	s.Equal(len(experiment.Runs), resp.RunCount)
}

func (s *GetExperimentTestSuite) Test_Error() {
	tests := []struct {
		name  string
		error string
		ID    string
	}{
		{
			name: "IncorrectExperimentID",
			error: `: unable to parse experiment id "incorrect_experiment_id": strconv.ParseInt: ` +
				`parsing "incorrect_experiment_id": invalid syntax`,
			ID: "incorrect_experiment_id",
		},
		{
			name:  "NotFoundExperiment",
			error: `: Not Found`,
			ID:    "1",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var resp api.ErrorResponse
			s.Require().Nil(s.AIMClient().WithResponse(&resp).DoRequest("/experiments/%s", tt.ID))
			s.Equal(tt.error, resp.Error())
		})
	}
}
