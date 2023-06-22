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
			name: "QueryWithMetricNameNoSpaces",
			query: `(run.active == true) and ((metric.name=="accuracy"))`,
			wantResult: true,
		},
		{
			name: "QueryWithStringSyntax",
			query: `(run.active == true) and (metric.name.startswith("acc"))`,
			wantResult: true,
		},
		{
			name: "QueryWithInSyntax",
			query: `(run.active == true) and ("accuracy" in metric.name)`,
			wantResult: true,
		},
		{
			name: "QueryWithoutMetricName",
			query: `(run.active == true)`,
			wantResult: false,
		},
		{
			name: "QueryWithTrickyDictKey",
			query: `(run.active == true) and (run.tags["metric.name"] == "foo")`,
			wantResult: false,
		},
		{
			name: "QueryWithTrickyDictKeyAndMetricName",
			query: `(run.active == true) and (run.tags["mymetric.name"] == "foo") and (metric.name == "accuracy")`,
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateMetricNamePresent(tt.query)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
