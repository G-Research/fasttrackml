package run

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
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
		ID:         strings.ReplaceAll(uuid.New().String(), "-", ""),
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
	for i := 0; i < 5; i++ {
		_, err = s.ArtifactFixtures.CreateArtifact(context.Background(), &models.Artifact{
			ID:      uuid.New(),
			Name:    "some-name",
			RunID:   run1.ID,
			BlobURI: "path/filename.png",
			Step:    int64(i),
			Iter:    1,
			Index:   1,
			Caption: "caption1",
			Format:  "png",
			Width:   100,
			Height:  100,
		})
		s.Require().Nil(err)
	}

	run2, err := s.RunFixtures.CreateRun(context.Background(), &models.Run{
		ID:         strings.ReplaceAll(uuid.New().String(), "-", ""),
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
	for i := 0; i < 5; i++ {
		_, err = s.ArtifactFixtures.CreateArtifact(context.Background(), &models.Artifact{
			ID:      uuid.New(),
			Name:    "other-name",
			RunID:   run2.ID,
			BlobURI: "path/filename.png",
			Step:    int64(i),
			Iter:    1,
			Index:   1,
			Caption: "caption2",
			Format:  "png",
			Width:   100,
			Height:  100,
		})
		s.Require().Nil(err)
	}

	tests := []struct {
		name                string
		request             request.SearchArtifactsRequest
		includedRuns        []*models.Run
		excludedRuns        []*models.Run
		expectedRecordRange int64
		expectedIndexRange  int64
	}{
		{
			name:                "SearchArtifact",
			request:             request.SearchArtifactsRequest{},
			includedRuns:        []*models.Run{run1, run2},
			expectedRecordRange: 5,
			expectedIndexRange:  1,
		},
		{
			name: "SearchArtifactWithNameQuery",
			request: request.SearchArtifactsRequest{
				Query: `((images.name == "some-name"))`,
			},
			includedRuns:        []*models.Run{run1},
			excludedRuns:        []*models.Run{run2},
			expectedRecordRange: 5,
			expectedIndexRange:  1,
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

			for _, run := range tt.includedRuns {
				traceIndex := 0
				imgIndex := 0
				valuesIndex := 0
				rangesPrefix := fmt.Sprintf("%v.ranges", run.ID)
				recordRangeKey := rangesPrefix + ".record_range_total.1"
				s.Equal(tt.expectedRecordRange, decodedData[recordRangeKey])
				indexRangeKey := rangesPrefix + ".index_range_total.1"
				s.Equal(tt.expectedIndexRange, decodedData[indexRangeKey])
				tracesPrefix := fmt.Sprintf("%v.traces.%d", run.ID, traceIndex)
				valuesPrefix := fmt.Sprintf(".values.%d.%d", valuesIndex, imgIndex)
				blobUriKey := tracesPrefix + valuesPrefix + ".blob_uri"
				s.Equal("path/filename.png", decodedData[blobUriKey])
			}
			for _, run := range tt.excludedRuns {
				imgIndex := 0
				valuesIndex := 0
				rangesPrefix := fmt.Sprintf("%v.ranges", run.ID)
				recordRangeKey := rangesPrefix + ".record_range_total.1"
				s.Empty(decodedData[recordRangeKey])
				indexRangeKey := rangesPrefix + ".index_range_total.1"
				s.Empty(decodedData[indexRangeKey])
				tracesPrefix := fmt.Sprintf("%v.traces.%d", run.ID, imgIndex)
				valuesPrefix := fmt.Sprintf(".values.%d", valuesIndex)
				blobUriKey := tracesPrefix + valuesPrefix + ".blob_uri"
				s.Empty(decodedData[blobUriKey])
			}
		})
	}
}
