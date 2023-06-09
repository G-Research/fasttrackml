package query

import (
	"testing"

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

func (s *QueryTestSuite) Test_Ok() {
	tests := []struct {
		name         string
		query        string
		expectedSQL  string
		expectedVars []interface{}
	}{
		{
			name:         "TestRunNameWithoutFunction",
			query:        `(run.name == 'run')`,
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" = $1 AND "runs"."lifecycle_stage" <> $2) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{"run", models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithContainsFunction",
			query:        `(run.name.contains('run'))`,
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" LIKE '%run%' AND "runs"."lifecycle_stage" <> $1) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithStartWithFunction",
			query:        `(run.name.startswith('run'))`,
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" LIKE 'run%' AND "runs"."lifecycle_stage" <> $1) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{models.LifecycleStageDeleted},
		},
		{
			name:         "TestRunNameWithEndWithFunction",
			query:        `(run.name.endswith('run'))`,
			expectedSQL:  `SELECT * FROM "runs" WHERE ("runs"."name" LIKE '%run' AND "runs"."lifecycle_stage" <> $1) ORDER BY "runs"."run_uuid" LIMIT 1`,
			expectedVars: []interface{}{models.LifecycleStageDeleted},
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
