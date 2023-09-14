package aim

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/aim/query"
)

func Test_isMetricSelected(t *testing.T) {
	dialectors := []string{
		"sqlite3",
		"postgres",
	}
	for _, dialector := range dialectors {
		qp := query.QueryParser{
			Default: query.DefaultExpression{
				Contains:   "run.archived",
				Expression: "not run.archived",
			},
			Tables: map[string]string{
				"runs":        "runs",
				"experiments": "experiments",
				"metrics":     "latest_metrics",
			},
			TzOffset:  0,
			Dialector: dialector,
		}

		tests := []struct {
			name       string
			query      string
			wantResult bool
		}{
			{
				name:       "QueryWithMetricName",
				query:      `(run.active == True) and ((metric.name == "accuracy"))`,
				wantResult: true,
			},
			{
				name:       "QueryWithMetricNameNoSpaces",
				query:      `(run.active == True) and ((metric.name=="accuracy"))`,
				wantResult: true,
			},
			{
				name:       "QueryWithStringSyntax",
				query:      `(run.active == True) and (metric.name.startswith("acc"))`,
				wantResult: true,
			},
			{
				name:       "QueryWithInSyntax",
				query:      `(run.active == True) and ("accuracy" in metric.name)`,
				wantResult: true,
			},
			{
				name:       "QueryWithMetricNameAtEnd",
				query:      `(run.active == True) and "accuracy" in metric.name`,
				wantResult: true,
			},
			{
				name:       "QueryWithoutMetricName",
				query:      `(run.active == True)`,
				wantResult: false,
			},
			{
				name:       "QueryWithTrickyDictKey",
				query:      `(run.active == True) and (run.tags["metric.name"] == "foo")`,
				wantResult: false,
			},
			{
				name:       "QueryWithTrickyDictKeyAndMetricName",
				query:      `(run.active == True) and (run.tags["mymetric.name"] == "foo") and (metric.name == "accuracy")`,
				wantResult: true,
			},
			{
				name:       "QueryWithRegexMatchMetricName",
				query:      `(run.active == True) and re.match("accuracy", metric.name)`,
				wantResult: true,
			},
			{
				name:       "QueryWithRegexSearchMetricName",
				query:      `(run.active == True) and re.search("accuracy", metric.name)`,
				wantResult: true,
			},
			{
				name:       "QueryWithRegexSearchAndNoMetricName",
				query:      `(run.active == True) and re.search("accuracy", run.tags["metric.name"])`,
				wantResult: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				pq, err := qp.Parse(tt.query)
				assert.Nil(t, err)
				assert.Equal(t, tt.wantResult, pq.IsMetricSelected())
			})
		}
	}
}
