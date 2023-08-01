package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/response"
)

// CompareExpectedSearchRunsResponseWithActualSearchRunsResponse compares
// expected response object with the response from POST /runs/search` endpoint.
func CompareExpectedSearchRunsResponseWithActualSearchRunsResponse(
	t *testing.T, expectedResponse *response.SearchRunsResponse, actualResponse *response.SearchRunsResponse,
) {
	assert.Equal(t, len(expectedResponse.Runs), len(actualResponse.Runs))
	assert.Equal(t, len(expectedResponse.NextPageToken), len(actualResponse.NextPageToken))

	mappedExpectedResult := make(map[string]*response.RunPartialResponse, len(expectedResponse.Runs))
	for _, run := range expectedResponse.Runs {
		mappedExpectedResult[run.Info.ID] = run
	}

	if actualResponse.Runs != nil && expectedResponse.Runs != nil {
		for _, actualRun := range actualResponse.Runs {
			expectedRun, ok := mappedExpectedResult[actualRun.Info.ID]
			assert.True(t, ok)
			assert.NotEmpty(t, actualRun.Info.ID)
			assert.Equal(t, expectedRun.Info.Name, actualRun.Info.Name)
			assert.Equal(t, expectedRun.Info.Name, actualRun.Info.Name)
			assert.Equal(t, expectedRun.Info.UserID, actualRun.Info.UserID)
			assert.Equal(t, expectedRun.Info.Status, actualRun.Info.Status)
			assert.Equal(t, expectedRun.Info.EndTime, actualRun.Info.EndTime)
			assert.Equal(t, expectedRun.Info.StartTime, actualRun.Info.StartTime)
			assert.Equal(t, expectedRun.Info.ArtifactURI, actualRun.Info.ArtifactURI)
			assert.Equal(t, expectedRun.Info.ExperimentID, actualRun.Info.ExperimentID)
			assert.Equal(t, expectedRun.Info.LifecycleStage, actualRun.Info.LifecycleStage)
		}
	}
}
