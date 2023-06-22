package aim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateMetricNamePresent(t *testing.T) {
	tests := []struct{
		name string
		query string
		wantResult bool
	}{
		{
			name: "QueryWithMetricName",
			query: `(run.active == true) and ((metric.name == "accuracy"))`,
			wantResult: true,
		},
		{
			name: "QueryWithoutMetricName",
			query: `(run.active == true)`,
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateMetricNamePresent(tt.query)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
