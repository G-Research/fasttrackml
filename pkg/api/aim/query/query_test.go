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

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
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
				`WHERE "runs"."name" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithInFunction",
			query: `('run' in run.name)`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"%run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNotInFunction",
			query: `('run' not in run.name)`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" NOT LIKE $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"%run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithStartWithFunction",
			query: `(run.name.startswith('run'))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithEndWithFunction",
			query: `(run.name.endswith('run'))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"%run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithRegexpMatchFunction",
			query: `(re.match('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" ~ $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"^run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithRegexpSearchFunction",
			query: `(re.search('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" ~ $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNegatedRegexpMatchFunction",
			query: `not (re.match('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" !~ $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"^run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNegatedRegexpSearchFunction",
			query: `not (re.search('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" !~ $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestNegativeInteger",
			query: `run.metrics['my_metric'].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`WHERE "metrics_0"."value" < $2 AND "runs"."lifecycle_stage" <> $3`,
			expectedVars: []interface{}{"my_metric", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestNegativeFloat",
			query: `run.metrics['my_metric'].last < -1.0`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`WHERE "metrics_0"."value" < $2 AND "runs"."lifecycle_stage" <> $3`,
			expectedVars: []interface{}{"my_metric", -1.0, models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSliceTuple",
			query: `run.metrics["my_metric", {"key1": "value1"}].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`LEFT JOIN contexts contexts_1 ON metrics_0.context_id = contexts_1.id ` +
				`WHERE "contexts_1"."json"#>>$2 = $3 ` +
				`AND ("metrics_0"."value" < $4 AND "runs"."lifecycle_stage" <> $5)`,
			expectedVars: []interface{}{"my_metric", "{key1}", "value1", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSliceTupleWithPrefix",
			query: `run.metrics["my_metric", {"$.key1": "value1"}].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`LEFT JOIN contexts contexts_1 ON metrics_0.context_id = contexts_1.id ` +
				`WHERE "contexts_1"."json"#>>$2 = $3 ` +
				`AND ("metrics_0"."value" < $4 AND "runs"."lifecycle_stage" <> $5)`,
			expectedVars: []interface{}{"my_metric", "{key1}", "value1", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestTagsSubscript",
			query: `(run.tags["foo"] == "bar")`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" LEFT JOIN tags tags_0 ON runs.run_uuid = tags_0.run_uuid ` +
				`AND tags_0.key = $1 WHERE "tags_0"."value" = $2 AND "runs"."lifecycle_stage" <> $3`,
			expectedVars: []interface{}{"foo", "bar", models.LifecycleStageDeleted},
		},
		{
			name:  "TestTagsAttribute",
			query: `(run.tags.foo == "bar")`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" LEFT JOIN tags tags_0 ON runs.run_uuid = tags_0.run_uuid ` +
				`AND tags_0.key = $1 WHERE "tags_0"."value" = $2 AND "runs"."lifecycle_stage" <> $3`,
			expectedVars: []interface{}{"foo", "bar", models.LifecycleStageDeleted},
		},
		{
			name:         "TestCreationTimeAttribute",
			query:        `run.creation_time == 12345678`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "runs"."start_time" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{12345678, models.LifecycleStageDeleted},
		},
		{
			name:         "TestCreatedAtAttribute",
			query:        `run.created_at == 12345678`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "runs"."start_time" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{12345678, models.LifecycleStageDeleted},
		},
		{
			name:         "TestEndTimeAttribute",
			query:        `run.end_time == 12345678`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "runs"."end_time" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{12345678, models.LifecycleStageDeleted},
		},
		{
			name:         "TestFinalizedAtAttribute",
			query:        `run.finalized_at == 12345678`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "runs"."end_time" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{12345678, models.LifecycleStageDeleted},
		},
		{
			name:         "TestHashAttribute",
			query:        `run.hash == "hash"`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "runs"."run_uuid" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"hash", models.LifecycleStageDeleted},
		},
		{
			name:         "TestNameAttribute",
			query:        `run.name == "run name"`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "runs"."name" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"run name", models.LifecycleStageDeleted},
		},
		{
			name:         "TestExperimentAttribute",
			query:        `run.experiment == "experiment name"`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "Experiment"."name" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"experiment name", models.LifecycleStageDeleted},
		},
		{
			name:         "TestArchivedAttribute",
			query:        `run.archived`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "runs"."lifecycle_stage" = $1`,
			expectedVars: []interface{}{models.LifecycleStageDeleted},
		},
		{
			name:         "TestActiveAttribute",
			query:        `run.active`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "runs"."status" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{models.StatusRunning, models.LifecycleStageDeleted},
		},
		{
			name:  "TestDurationAttribute",
			query: `run.duration == 123456789`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" WHERE (runs.end_time - runs.start_time) / 1000 = $1 AND ` +
				`"runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{123456789, models.LifecycleStageDeleted},
		},
		{
			name:         "TestDatetimeFunction",
			query:        `run.creation_time > datetime(2022, 2, 2)`,
			expectedSQL:  `SELECT "run_uuid" FROM "runs" WHERE "runs"."start_time" > $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{int64(1643760000000), models.LifecycleStageDeleted},
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
				`WHERE "runs"."name" = $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithInFunction",
			query: `('run' in run.name)`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"%run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNotInFunction",
			query: `('run' not in run.name)`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" NOT LIKE $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"%run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithStartWithFunction",
			query: `(run.name.startswith('run'))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"run%", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithEndWithFunction",
			query: `(run.name.endswith('run'))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE "runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"%run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithRegexpMatchFunction",
			query: `(re.match('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE IFNULL("runs"."name", '') REGEXP $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"^run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithRegexpSearchFunction",
			query: `(re.search('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE IFNULL("runs"."name", '') REGEXP $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNegatedRegexpMatchFunction",
			query: `not (re.match('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE IFNULL("runs"."name", '') NOT REGEXP $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"^run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestRunNameWithNegatedRegexpSearchFunction",
			query: `not (re.search('run', run.name))`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`WHERE IFNULL("runs"."name", '') NOT REGEXP $1 AND "runs"."lifecycle_stage" <> $2`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:  "TestNegativeInteger",
			query: `run.metrics['my_metric'].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`WHERE "metrics_0"."value" < $2 AND "runs"."lifecycle_stage" <> $3`,
			expectedVars: []interface{}{"my_metric", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestNegativeFloat",
			query: `run.metrics['my_metric'].last < -1.0`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`WHERE "metrics_0"."value" < $2 AND "runs"."lifecycle_stage" <> $3`,
			expectedVars: []interface{}{"my_metric", -1.0, models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricKeySlice",
			query: `run.metrics["key1"].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`WHERE "metrics_0"."value" < $2 AND "runs"."lifecycle_stage" <> $3`,
			expectedVars: []interface{}{"key1", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSliceTuple",
			query: `run.metrics["my_metric", {"key1": "value1"}].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`LEFT JOIN contexts contexts_1 ON metrics_0.context_id = contexts_1.id ` +
				`WHERE IFNULL("contexts_1"."json", JSON('{}'))->>$2 = $3 ` +
				`AND ("metrics_0"."value" < $4 AND "runs"."lifecycle_stage" <> $5)`,
			expectedVars: []interface{}{"my_metric", "$.key1", "value1", -1, models.LifecycleStageDeleted},
		},
		{
			name:  "TestMetricContextSliceTupleWithPrefix",
			query: `run.metrics["my_metric", {"$.key1": "value1"}].last < -1`,
			expectedSQL: `SELECT "run_uuid" FROM "runs" ` +
				`LEFT JOIN latest_metrics metrics_0 ON runs.run_uuid = metrics_0.run_uuid AND metrics_0.key = $1 ` +
				`LEFT JOIN contexts contexts_1 ON metrics_0.context_id = contexts_1.id ` +
				`WHERE IFNULL("contexts_1"."json", JSON('{}'))->>$2 = $3 ` +
				`AND ("metrics_0"."value" < $4 AND "runs"."lifecycle_stage" <> $5)`,
			expectedVars: []interface{}{"my_metric", "$.key1", "value1", -1, models.LifecycleStageDeleted},
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
			name:          "TestMetricContextSubscriptTupleWrongOrder",
			query:         `run.metrics[{"key1": "value1"}, "my_metric"].last < -1`,
			expectedError: SyntaxError{},
		},
		{
			name:          "TestMetricContextSubscriptTupleDictOnly",
			query:         `run.metrics[{"key1": "value1"}].last < -1`,
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
