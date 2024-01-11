package query

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

type QueryTestSuite struct {
	db *gorm.DB
	suite.Suite
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

func (s *QueryTestSuite) SetupTest() {
	mockedDB, _, err := sqlmock.New()
	require.Nil(s.T(), err)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:       mockedDB,
		DriverName: "postgres",
	}), &gorm.Config{})
	require.Nil(s.T(), err)
	s.db = db
}

func (s *QueryTestSuite) TestPostgresDialector_Ok() {
	tests := []struct {
		name          string
		query         string
		selectMetrics bool
		expectedSQL   string
		expectedVars  []interface{}
	}{
		{
			name:  "TestRunNameWithoutFunction",
			query: `(run.name == 'run')`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" = $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithInFunction",
			query: `('run' in run.name)`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"%run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNotInFunction",
			query: `('run' not in run.name)`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" NOT LIKE $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"%run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithStartWithFunction",
			query: `(run.name.startswith('run'))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithEndWithFunction",
			query: `(run.name.endswith('run'))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"%run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithRegexpMatchFunction",
			query: `(re.match('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" ~ $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"^run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithRegexpSearchFunction",
			query: `(re.search('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" ~ $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNegatedRegexpMatchFunction",
			query: `not (re.match('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" !~ $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"^run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNegatedRegexpSearchFunction",
			query: `not (re.search('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" !~ $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestNegativeInteger",
			query: `run.metrics['my_metric'].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics ON latest_metrics.run_uuid = runs.run_uuid ` +
				`WHERE ("latest_metrics"."key" = $1 AND ("latest_metrics"."value" < $2 AND "runs"."lifecycle_stage" <> $3))`,
			expectedVars: []interface{}{"my_metric", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestNegativeFloat",
			query: `run.metrics['my_metric'].last < -1.0`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`WHERE ("metrics_0"."value" < $2 AND "runs"."lifecycle_stage" <> $3)`,
			expectedVars: []interface{}{"my_metric", -1.0, models.LifecycleStageDeleted},
		},
		{
			name:          "TestMetricContext",
			query:         `metric.context.key1 == 'value1'`,
			selectMetrics: true,
			expectedSQL: `SELECT ID FROM "metrics" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`WHERE ("contexts"."json"#>>$1 = $2 AND "runs"."lifecycle_stage" <> $3)`,
			expectedVars: []interface{}{"{key1}", "value1", models.LifecycleStageDeleted},
		},
		{
			name:          "TestMetricContextNegative",
			query:         `metric.context.key1 != 'value1'`,
			selectMetrics: true,
			expectedSQL: `SELECT ID FROM "metrics" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`WHERE ("contexts"."json"#>>$1 <> $2 AND "runs"."lifecycle_stage" <> $3)`,
			expectedVars: []interface{}{"{key1}", "value1", models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSlice",
			query: `run.metrics[{"key1": "value1"}].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`WHERE ("contexts"."json"#>>$1 = $2 AND ("latest_metrics"."value" < $3 ` +
				`AND "runs"."lifecycle_stage" <> $4))`,
			expectedVars: []interface{}{"{key1}", "value1", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSliceTuple1",
			query: `run.metrics["my_metric", {"key1": "value1"}].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`LEFT JOIN latest_metrics metrics_1 ON runs.run_uuid = metrics_1.run_uuid AND metrics_1.key = $1 ` +
				`WHERE ("contexts"."json"#>>$2 = $3 AND ` +
				`("latest_metrics"."value" < $4 AND "runs"."lifecycle_stage" <> $5))`,
			expectedVars: []interface{}{"my_metric", "{key1}", "value1", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSliceTuple2",
			query: `run.metrics[{"key1": "value1"}, "my_metric"].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`LEFT JOIN latest_metrics metrics_1 ON runs.run_uuid = metrics_1.run_uuid AND metrics_1.key = $1 ` +
				`WHERE ("contexts"."json"#>>$2 = $3 AND ` +
				`("metrics_1"."value" < $4 AND "runs"."lifecycle_stage" <> $5))`,
			expectedVars: []interface{}{"my_metric", "{key1}", "value1", -1, models.LifecycleStageDeleted},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			pq := QueryParser{
				Default: DefaultExpression{
					Contains:   "run.archived",
					Expression: "not run.archived",
				},
				Tables: map[string]string{
					"runs":        "runs",
					"experiments": "Experiment",
					"metrics":     "metrics",
				},
				Dialector: postgres.Dialector{}.Name(),
			}
			parsedQuery, err := pq.Parse(tt.query)
			require.Nil(s.T(), err)
			var tx *gorm.DB
			if tt.selectMetrics {
				tx = parsedQuery.Filter(
					s.db.Session(&gorm.Session{DryRun: true}).Model(models.Metric{}),
				).Select("ID").Find(models.Metric{})
			} else {
				tx = parsedQuery.Filter(
					s.db.Session(&gorm.Session{DryRun: true}).Model(models.Run{}),
				).Select("ID").Find(&models.Run{})
			}

			require.Nil(s.T(), tx.Error)
			assert.Equal(s.T(), tt.expectedSQL, tx.Statement.SQL.String())
			assert.Equal(s.T(), tt.expectedVars, tx.Statement.Vars)
		})
	}
}

func (s *QueryTestSuite) TestSqliteDialector_Ok() {
	tests := []struct {
		name          string
		query         string
		selectMetrics bool
		expectedSQL   string
		expectedVars  []interface{}
	}{
		{
			name:  "TestRunNameWithoutFunction",
			query: `(run.name == 'run')`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" = $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithInFunction",
			query: `('run' in run.name)`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"%run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNotInFunction",
			query: `('run' not in run.name)`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" NOT LIKE $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"%run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithStartWithFunction",
			query: `(run.name.startswith('run'))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithEndWithFunction",
			query: `(run.name.endswith('run'))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"%run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithRegexpMatchFunction",
			query: `(re.match('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE (IFNULL("runs"."name", '') REGEXP $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"^run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithRegexpSearchFunction",
			query: `(re.search('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE (IFNULL("runs"."name", '') REGEXP $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNegatedRegexpMatchFunction",
			query: `not (re.match('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE (IFNULL("runs"."name", '') NOT REGEXP $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"^run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNegatedRegexpSearchFunction",
			query: `not (re.search('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE (IFNULL("runs"."name", '') NOT REGEXP $1 AND "runs"."lifecycle_stage" <> $2)`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestNegativeInteger",
			query: `run.metrics['my_metric'].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`WHERE ("metrics_0"."value" < $2 AND "runs"."lifecycle_stage" <> $3)`,
			expectedVars: []interface{}{"my_metric", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestNegativeFloat",
			query: `run.metrics['my_metric'].last < -1.0`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`WHERE ("metrics_0"."value" < $2 AND "runs"."lifecycle_stage" <> $3)`,
			expectedVars: []interface{}{"my_metric", -1.0, models.LifecycleStageDeleted},
		},
		{
			name:          "TestMetricContext",
			query:         `metric.context.key1 == 'value1'`,
			selectMetrics: true,
			expectedSQL: `SELECT ID FROM "metrics" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`WHERE (IFNULL("contexts"."json", JSON('{}'))->>$1 = $2 AND "runs"."lifecycle_stage" <> $3)`,
			expectedVars: []interface{}{"key1", "value1", models.LifecycleStageDeleted},
		},
		{
			name:          "TestMetricContextNegative",
			query:         `metric.context.key1 != 'value1'`,
			selectMetrics: true,
			expectedSQL: `SELECT ID FROM "metrics" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`WHERE (IFNULL("contexts"."json", JSON('{}'))->>$1 <> $2 AND "runs"."lifecycle_stage" <> $3)`,
			expectedVars: []interface{}{"key1", "value1", models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSlice",
			query: `run.metrics[{"key1": "value1"}].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`WHERE (IFNULL("contexts"."json", JSON('{}'))->>$1 = $2 AND ` +
				`("latest_metrics"."value" < $3 AND "runs"."lifecycle_stage" <> $4))`,
			expectedVars: []interface{}{"key1", "value1", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSliceTuple1",
			query: `run.metrics["my_metric", {"key1": "value1"}].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`LEFT JOIN latest_metrics metrics_1 ON runs.run_uuid = metrics_1.run_uuid AND metrics_1.key = $1 ` +
				`WHERE (IFNULL("contexts"."json", JSON('{}'))->>$2 = $3 AND ` +
				`("latest_metrics"."value" < $4 AND "runs"."lifecycle_stage" <> $5))`,
			expectedVars: []interface{}{"my_metric", "key1", "value1", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSliceTuple2",
			query: `run.metrics[{"key1": "value1"}, "my_metric"].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN contexts ON latest_metrics.context_id = contexts.id ` +
				`LEFT JOIN latest_metrics metrics_1 ON runs.run_uuid = metrics_1.run_uuid AND metrics_1.key = $1 ` +
				`WHERE (IFNULL("contexts"."json", JSON('{}'))->>$2 = $3 AND ` +
				`("metrics_1"."value" < $4 AND "runs"."lifecycle_stage" <> $5))`,
			expectedVars: []interface{}{"my_metric", "key1", "value1", -1, models.LifecycleStageDeleted},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			pq := QueryParser{
				Default: DefaultExpression{
					Contains:   "run.archived",
					Expression: "not run.archived",
				},
				Tables: map[string]string{
					"runs":        "runs",
					"experiments": "Experiment",
					"metrics":     "metrics",
				},
				Dialector: sqlite.Dialector{}.Name(),
			}
			parsedQuery, err := pq.Parse(tt.query)
			require.Nil(s.T(), err)
			var tx *gorm.DB
			if tt.selectMetrics {
				tx = parsedQuery.Filter(
					s.db.Session(&gorm.Session{DryRun: true}).Model(models.Metric{}),
				).Select("ID").Find(models.Metric{})
			} else {
				tx = parsedQuery.Filter(
					s.db.Session(&gorm.Session{DryRun: true}).Model(models.Run{}),
				).Select("ID").Find(&models.Run{})
			}

			require.Nil(s.T(), tx.Error)
			assert.Equal(s.T(), tt.expectedSQL, tx.Statement.SQL.String())
			assert.Equal(s.T(), tt.expectedVars, tx.Statement.Vars)
		})
	}
}

func (s *QueryTestSuite) Test_Error() {
	tests := []struct {
		name          string
		query         string
		expectedError error
	}{
		{
			name:          "TestMetricContextNested",
			query:         `metric.context.parent.nested == 'value1'`,
			expectedError: SyntaxError{},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			pq := QueryParser{
				Default: DefaultExpression{
					Contains:   "run.archived",
					Expression: "not run.archived",
				},
				Tables: map[string]string{
					"runs":        "runs",
					"experiments": "Experiment",
					"metrics":     "metrics",
				},
				Dialector: sqlite.Dialector{}.Name(),
			}
			parsedQuery, err := pq.Parse(tt.query)
			require.IsType(s.T(), tt.expectedError, err)
			require.Nil(s.T(), parsedQuery)
		})
	}
}
