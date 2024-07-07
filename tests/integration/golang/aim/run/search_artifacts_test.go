package run

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/encoding"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type SearchArtifactsTestSuite struct {
	helpers.BaseTestSuite
}

func TestSearchArtifactsTestSuite(t *testing.T) {
	suite.Run(t, new(SearchArtifactsTestSuite))
}

func (s *SearchArtifactsTestSuite) Test_Ok() {
	// create test experiments.
	experiment, err := s.ExperimentFixtures.CreateExperiment(context.Background(), &models.Experiment{
		Name:           uuid.New().String(),
		LifecycleStage: models.LifecycleStageActive,
		NamespaceID:    s.DefaultNamespace.ID,
	})
	s.Require().Nil(err)

	// create different test runs and attach tags, metrics, params, etc.
	run1, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id1",
		Name:       "TestRun1",
		UserID:     "1",
		Status:     models.StatusRunning,
		SourceType: "JOB",
		StartTime: sql.NullInt64{
			Int64: 123456789,
			Valid: true,
		},
		EndTime: sql.NullInt64{
			Int64: 123456789,
			Valid: true,
		},
		ExperimentID:   *experiment.ID,
		ArtifactURI:    "artifact_uri1",
		LifecycleStage: models.LifecycleStageActive,
	})
	s.Require().Nil(err)
	_, err = s.MetricFixtures.CreateMetric(context.Background(), &models.Metric{
		Key:       "TestMetric1",
		Value:     1.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run1.ID,
		Iter:      1,
	})
	s.Require().Nil(err)
	metric1Run1, err := s.MetricFixtures.CreateLatestMetric(context.Background(), &models.LatestMetric{
		Key:       "TestMetric1",
		Value:     1.1,
		Timestamp: 123456789,
		Step:      1,
		IsNan:     false,
		RunID:     run1.ID,
		LastIter:  1,
	})
	s.Require().Nil(err)

	runs := []*models.Run{run1}
	tests := []struct {
		name    string
		request request.SearchArtifactsRequest
		metrics []*models.LatestMetric
	}{
		{
			name:    "SearchArtifact",
			request: request.SearchArtifactsRequest{},
			metrics: []*models.LatestMetric{
				metric1Run1,
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := new(bytes.Buffer)
			s.Require().Nil(
				s.AIMClient().WithMethod(
					http.MethodPost,
				).WithRequest(
					tt.request,
				).WithResponseType(
					helpers.ResponseTypeBuffer,
				).WithResponse(
					resp,
				).DoRequest("/runs/search/images"),
			)

			decodedData, err := encoding.NewDecoder(resp).Decode()
			s.Require().Nil(err)

			var decodedMetrics []*models.LatestMetric
			for _, run := range runs {
				metricCount := 0
				for decodedData[fmt.Sprintf("%v.traces.%d.name", run.ID, metricCount)] != nil {
					prefix := fmt.Sprintf("%v.traces.%d", run.ID, metricCount)
					epochsKey := prefix + ".epochs.blob"
					itersKey := prefix + ".iters.blob"
					nameKey := prefix + ".name"
					timestampsKey := prefix + ".timestamps.blob"
					valuesKey := prefix + ".values.blob"

					contextPrefix := prefix + ".context"
					contx, err := helpers.ExtractContextBytes(contextPrefix, decodedData)
					s.Require().Nil(err)

					decodedContext, err := s.ContextFixtures.GetContextByJSON(
						context.Background(),
						string(contx),
					)
					s.Require().Nil(err)

					m := models.LatestMetric{
						Key:       decodedData[nameKey].(string),
						Value:     decodedData[valuesKey].([]float64)[0],
						Timestamp: int64(decodedData[timestampsKey].([]float64)[0] * 1000),
						Step:      int64(decodedData[epochsKey].([]float64)[0]),
						IsNan:     false,
						RunID:     run.ID,
						LastIter:  int64(decodedData[itersKey].([]float64)[0]),
						ContextID: decodedContext.ID,
						Context:   *decodedContext,
					}
					decodedMetrics = append(decodedMetrics, &m)
					metricCount++
				}
			}
			// Check if the received metrics match the expected ones
			s.Equal(len(tt.metrics), len(decodedMetrics))
			for i, metric := range tt.metrics {
				s.Equal(metric, decodedMetrics[i])
			}
		})
	}
}
