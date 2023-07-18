package query

import (
	"testing"

	"gorm.io/driver/sqlite"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
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
	assert.Nil(s.T(), err)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:       mockedDB,
		DriverName: "postgres",
	}), &gorm.Config{})
	assert.Nil(s.T(), err)
	s.db = db
}

func (s *QueryTestSuite) TestPostgresDialector_Ok() {
	tests := []struct {
		name         string
		query        string
		dialector    string
		expectedSQL  string
		expectedVars []interface{}
	}{
		{
			name:         "TestRunNameWithoutFunction",
			query:        `(run.name == 'run')`,
			dialector:    postgres.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" = $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithContainsFunction",
			query:        `(run.name.contains('run'))`,
			dialector:    postgres.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"'%run%'", models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithStartWithFunction",
			query:        `(run.name.startswith('run'))`,
			dialector:    postgres.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"'run%'", models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithEndWithFunction",
			query:        `(run.name.endswith('run'))`,
			dialector:    postgres.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"'%run'", models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithMatchFunction",
			query:        `(run.name.match('run'))`,
			dialector:    postgres.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" ~ $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"'run'", models.LifecycleStageDeleted},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			pq := QueryParser{
				Default: DefaultExpression{
					Contains:   "run.archived",
					Expression: "not run.archived",
				},
				Tables: map[string]string{
					"runs":        "runs",
					"experiments": "Experiment",
				},
				Dialector: tt.dialector,
			}
			parsedQuery, err := pq.Parse(tt.query)
			assert.Nil(s.T(), err)
			result := parsedQuery.Filter(
				s.db.Session(&gorm.Session{DryRun: true}).Model(models.Run{}),
			).First(&models.Run{})
			assert.Nil(s.T(), result.Error)
			assert.Equal(s.T(), tt.expectedSQL, result.Statement.SQL.String())
			assert.Equal(s.T(), tt.expectedVars, result.Statement.Vars)
		})
	}
}

func (s *QueryTestSuite) TestSqliteDialector_Ok() {
	tests := []struct {
		name         string
		query        string
		dialector    string
		expectedSQL  string
		expectedVars []interface{}
	}{
		{
			name:         "TestRunNameWithoutFunction",
			query:        `(run.name == 'run')`,
			dialector:    sqlite.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" = $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithContainsFunction",
			query:        `(run.name.contains('run'))`,
			dialector:    sqlite.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"'%run%'", models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithStartWithFunction",
			query:        `(run.name.startswith('run'))`,
			dialector:    sqlite.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"'run%'", models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithEndWithFunction",
			query:        `(run.name.endswith('run'))`,
			dialector:    sqlite.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" LIKE $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"'%run'", models.LifecycleStageDeleted},
		},
		{
			name:         "SqliteDialector/TestRunNameWithMatchFunction",
			query:        `(run.name.match('run'))`,
			dialector:    sqlite.Dialector{}.Name(),
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" regexp $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"'run'", models.LifecycleStageDeleted},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(T *testing.T) {
			pq := QueryParser{
				Default: DefaultExpression{
					Contains:   "run.archived",
					Expression: "not run.archived",
				},
				Tables: map[string]string{
					"runs":        "runs",
					"experiments": "Experiment",
				},
				Dialector: tt.dialector,
			}
			parsedQuery, err := pq.Parse(tt.query)
			assert.Nil(s.T(), err)
			result := parsedQuery.Filter(
				s.db.Session(&gorm.Session{DryRun: true}).Model(models.Run{}),
			).First(&models.Run{})
			assert.Nil(s.T(), result.Error)
			assert.Equal(s.T(), tt.expectedSQL, result.Statement.SQL.String())
			assert.Equal(s.T(), tt.expectedVars, result.Statement.Vars)
		})
	}
}

func (s *QueryTestSuite) Test_Error() {}
