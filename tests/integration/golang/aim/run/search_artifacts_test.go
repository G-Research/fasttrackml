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
	_, err = s.ArtifactFixtures.CreateArtifact(context.Background(), &models.Artifact{
		ID:      uuid.New(),
		RunID:   run1.ID,
		BlobURI: "path/filename.png",
		Step:    1,
		Iter:    1,
		Index:   1,
		Caption: "caption1",
		Format:  "png",
		Width:   100,
		Height:  100,
	})
	s.Require().Nil(err)

	run2, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:         "id2",
		Name:       "TestRun2",
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
	_, err = s.ArtifactFixtures.CreateArtifact(context.Background(), &models.Artifact{
		ID:      uuid.New(),
		RunID:   run2.ID,
		BlobURI: "path/filename.png",
		Step:    1,
		Iter:    1,
		Index:   1,
		Caption: "caption2",
		Format:  "png",
		Width:   100,
		Height:  100,
	})
	s.Require().Nil(err)

	runs := []*models.Run{run1, run2}
	tests := []struct {
		name    string
		request request.SearchArtifactsRequest
		metrics []*models.LatestMetric
	}{
		{
			name:    "SearchArtifact",
			request: request.SearchArtifactsRequest{},
			metrics: []*models.LatestMetric{},
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
				).DoRequest("/runs/search/image"),
			)

			decodedData, err := encoding.NewDecoder(resp).Decode()
			s.Require().Nil(err)

			for _, run := range runs {
				imgIndex := 0
				rangesPrefix := fmt.Sprintf("%v.ranges", run.ID)
				recordRangeKey := rangesPrefix + ".record_range_total.1"
				s.Equal(int64(1), decodedData[recordRangeKey])
				indexRangeKey := rangesPrefix + ".index_range_total.1"
				s.Equal(int64(1), decodedData[indexRangeKey])
				tracesPrefix := fmt.Sprintf("%v.traces.%d", run.ID, imgIndex)
				blobUriKey := tracesPrefix + ".blob_uri"
				s.Equal("path/filename.png", decodedData[blobUriKey])
			}
		})
	}
}
