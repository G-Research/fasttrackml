package convertors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
)

func TestConvertMetricParamRequestToDBModel(t *testing.T) {
	req := request.LogMetricRequest{
		Key:       "key",
		Step:      1,
		Value:     1.1,
		RunID:     "run_id",
		Timestamp: 1234567890,
	}
	result, err := ConvertMetricParamRequestToDBModel("run_id", nil, &req)
	require.Nil(t, err)
	assert.Equal(t, "key", result.Key)
	assert.Equal(t, int64(1), result.Step)
	assert.Equal(t, 1.1, result.Value)
	assert.Equal(t, "run_id", result.RunID)
	assert.Equal(t, int64(1234567890), result.Timestamp)
}
