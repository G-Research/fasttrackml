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
		Name:             uuid.New().String(),
		LifecycleStage:   models.LifecycleStageActive,
		NamespaceID:      s.DefaultNamespace.ID,
		ArtifactLocation: "s3://my-bucket",
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
		for j := 0; j < 5; j++ {
			_, err = s.ArtifactFixtures.CreateArtifact(context.Background(), &models.Artifact{
				ID:      uuid.New(),
				Name:    "some-name",
				RunID:   run1.ID,
				BlobURI: "path/filename.png",
				Step:    int64(i),
				Iter:    1,
				Index:   int64(j),
				Caption: "caption1",
				Format:  "png",
				Width:   100,
				Height:  100,
			})
			s.Require().Nil(err)
		}
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
		for j := 0; j < 5; j++ {
			_, err = s.ArtifactFixtures.CreateArtifact(context.Background(), &models.Artifact{
				ID:      uuid.New(),
				Name:    "other-name",
				RunID:   run2.ID,
				BlobURI: "path/filename.png",
				Step:    int64(i),
				Iter:    1,
				Index:   int64(j),
				Caption: "caption2",
				Format:  "png",
				Width:   100,
				Height:  100,
			})
			s.Require().Nil(err)
		}
	}

	tests := []struct {
		name                         string
		request                      request.SearchArtifactsRequest
		includedRuns                 []*models.Run
		excludedRuns                 []*models.Run
		expectedRecordRangeUsedMax   int64
		expectedIndexRangeUsedMax    int64
		expectedImageIndexesPresent  []int
		expectedImageIndexesAbsent   []int
		expectedValuesIndexesPresent []int
		expectedValuesIndexesAbsent  []int
	}{
		{
			name: "SearchArtifact",
			request: request.SearchArtifactsRequest{
				Query: `((images.name == "some-name") or (images.name == "other-name"))`,
			},
			includedRuns:                 []*models.Run{run1, run2},
			expectedRecordRangeUsedMax:   4,
			expectedIndexRangeUsedMax:    4,
			expectedImageIndexesPresent:  []int{0, 1, 2, 3},
			expectedImageIndexesAbsent:   []int{},
			expectedValuesIndexesPresent: []int{0, 1, 2, 3, 4},
			expectedValuesIndexesAbsent:  []int{},
		},
		{
			name: "SearchArtifactWithNameQuery",
			request: request.SearchArtifactsRequest{
				Query: `((images.name == "some-name"))`,
			},
			includedRuns:                 []*models.Run{run1},
			excludedRuns:                 []*models.Run{run2},
			expectedRecordRangeUsedMax:   4,
			expectedIndexRangeUsedMax:    4,
			expectedImageIndexesPresent:  []int{0, 1, 2, 3},
			expectedImageIndexesAbsent:   []int{},
			expectedValuesIndexesPresent: []int{0, 1, 2, 3, 4},
			expectedValuesIndexesAbsent:  []int{},
		},
		{
			name: "SearchArtifactWithRecordRange",
			request: request.SearchArtifactsRequest{
				Query:       `((images.name == "some-name"))`,
				RecordRange: "0:2",
			},
			includedRuns:                 []*models.Run{run1},
			expectedRecordRangeUsedMax:   2,
			expectedIndexRangeUsedMax:    4,
			expectedImageIndexesPresent:  []int{0, 1, 2, 3},
			expectedImageIndexesAbsent:   []int{},
			expectedValuesIndexesPresent: []int{0, 1, 2},
			expectedValuesIndexesAbsent:  []int{3, 4},
		},
		{
			name: "SearchArtifactWithIndexRange",
			request: request.SearchArtifactsRequest{
				Query:      `((images.name == "some-name"))`,
				IndexRange: "0:2",
			},
			includedRuns:                 []*models.Run{run1},
			expectedRecordRangeUsedMax:   4,
			expectedIndexRangeUsedMax:    2,
			expectedImageIndexesPresent:  []int{0, 1, 2},
			expectedImageIndexesAbsent:   []int{3},
			expectedValuesIndexesPresent: []int{0, 1, 2, 3, 4},
			expectedValuesIndexesAbsent:  []int{},
		},
		{
			name: "SearchArtifactWithIndexDensity",
			request: request.SearchArtifactsRequest{
				Query:        `((images.name == "other-name"))`,
				IndexDensity: 1,
			},
			includedRuns:                 []*models.Run{run2},
			expectedRecordRangeUsedMax:   4,
			expectedIndexRangeUsedMax:    4,
			expectedImageIndexesPresent:  []int{0},
			expectedImageIndexesAbsent:   []int{1, 2, 3},
			expectedValuesIndexesPresent: []int{0, 1, 2, 3, 4},
			expectedValuesIndexesAbsent:  []int{},
		},
		{
			name: "SearchArtifactWithRecordDensity",
			request: request.SearchArtifactsRequest{
				Query:         `((images.name == "other-name"))`,
				RecordDensity: 1,
			},
			includedRuns:                 []*models.Run{run2},
			expectedRecordRangeUsedMax:   4,
			expectedIndexRangeUsedMax:    4,
			expectedImageIndexesPresent:  []int{0},
			expectedImageIndexesAbsent:   []int{1, 2, 3},
			expectedValuesIndexesPresent: []int{0},
			expectedValuesIndexesAbsent:  []int{1, 2, 3, 4},
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
				rangesPrefix := fmt.Sprintf("%v.ranges", run.ID)
				recordRangeKey := rangesPrefix + ".record_range_used.1"
				s.Equal(tt.expectedRecordRangeUsedMax, decodedData[recordRangeKey])
				propsPrefix := fmt.Sprintf("%v.props", run.ID)
				artifactLocation := propsPrefix + ".experiment.artifact_location"
				s.Equal(experiment.ArtifactLocation, decodedData[artifactLocation])
				indexRangeKey := rangesPrefix + ".index_range_used.1"
				s.Equal(tt.expectedIndexRangeUsedMax, decodedData[indexRangeKey])
				tracesPrefix := fmt.Sprintf("%v.traces.%d", run.ID, traceIndex)
				for _, valuesIndex := range tt.expectedValuesIndexesPresent {
					for _, imgIndex := range tt.expectedImageIndexesPresent {
						valuesPrefix := fmt.Sprintf(".values.%d.%d", valuesIndex, imgIndex)
						blobUriKey := tracesPrefix + valuesPrefix + ".blob_uri"
						s.Contains(decodedData, blobUriKey)
						s.Equal("path/filename.png", decodedData[blobUriKey])
					}
				}
				for _, valuesIndex := range tt.expectedValuesIndexesAbsent {
					for _, imgIndex := range tt.expectedImageIndexesAbsent {
						valuesPrefix := fmt.Sprintf(".values.%d.%d", valuesIndex, imgIndex)
						blobUriKey := tracesPrefix + valuesPrefix + ".blob_uri"
						s.NotContains(decodedData, blobUriKey)
					}
				}
			}
			for _, run := range tt.excludedRuns {
				imgIndex := 0
				valuesIndex := 0
				rangesPrefix := fmt.Sprintf("%v.ranges", run.ID)
				recordRangeKey := rangesPrefix + ".record_range_used.1"
				s.Empty(decodedData[recordRangeKey])
				propsPrefix := fmt.Sprintf("%v.props", run.ID)
				artifactLocation := propsPrefix + ".experiment.artifact_location"
				s.Empty(decodedData[artifactLocation])
				indexRangeKey := rangesPrefix + ".index_range_used.1"
				s.Empty(decodedData[indexRangeKey])
				tracesPrefix := fmt.Sprintf("%v.traces.%d", run.ID, imgIndex)
				valuesPrefix := fmt.Sprintf(".values.%d", valuesIndex)
				blobUriKey := tracesPrefix + valuesPrefix + ".blob_uri"
				s.Empty(decodedData[blobUriKey])
			}
		})
	}
}
