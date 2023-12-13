package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type JsonTestSuite struct {
	suite.Suite
	B  *testing.B
	db *sql.DB
}

func (s *JsonTestSuite) SetupSuite() {
	// setup sqlite Jsontest database from the schema
	jsontestDBPath := path.Join(s.T().TempDir(), "jsontest.db")
	db, err := sql.Open("sqlite3", jsontestDBPath)
	s.Require().Nil(err)
	s.db = db

	//nolint:gosec
	jsontestSql, err := os.ReadFile("jsondocschema.sql")
	s.Require().Nil(err)

	_, err = s.db.Exec(string(jsontestSql))
	s.Require().Nil(err)

	// Begin a transaction
	tx, err := s.db.Begin()
	s.Require().Nil(err)

	// Prepare a statement for inserting data
	contextStmt, err := tx.Prepare("INSERT INTO contexts(json) VALUES(?)")
	s.Require().Nil(err)
	defer contextStmt.Close() // Close the statement when we're done with it

	// Prepare a statement for inserting data into the 'metrics' table
	stmtMetrics, err := tx.Prepare("INSERT INTO metrics(key, value, context_id, context_null_id) VALUES(?, ?, ?, ?)")
	s.Require().Nil(err)
	defer stmtMetrics.Close() // Close the statement when we're done with it

	// Create a default/empty json context
	result, err := contextStmt.Exec("{}") // Insert the JSON document
	s.Require().Nil(err)
	defaultContextId, err := result.LastInsertId()
	s.Require().Nil(err)

	// Insert a large number of rows
	for i := 0; i < 1000000; i++ {
		// Create a JSON document with small variations
		jsonDoc := fmt.Sprintf(`{"key": "key%d", "value": "value%d"}`, i, i)

		result, err := contextStmt.Exec(jsonDoc) // Insert the JSON document
		s.Require().Nil(err)

		id, err := result.LastInsertId()
		s.Require().Nil(err)

		// Randomly decide whether to insert a null and default
		contextNullId := sql.NullInt64{Int64: id, Valid: true}
		if rand.Intn(2) == 0 { // 50% chance of being true
			contextNullId = sql.NullInt64{Int64: 0, Valid: false}
			id = defaultContextId
		}

		// Insert a row into the 'metrics' table that references the 'contexts' row
		key := fmt.Sprintf("key%d", i)
		value := rand.Float64()
		_, err = stmtMetrics.Exec(key, value, id, contextNullId) // Replace 'i' with the actual value you want to insert
		s.Require().Nil(err)
	}

	// Commit the transaction
	err = tx.Commit()
	s.Require().Nil(err)
}

func (s *JsonTestSuite) TearDownSuite() {
	// Close the database connection
	s.Require().Nil(s.db.Close())
}

func TestJsonTestSuite(t *testing.T) {
	suite.Run(t, new(JsonTestSuite))
}

func (s *JsonTestSuite) TestJson() {
	tests := []struct {
		name       string
		joinColumn string
		key        string
		value      string
	}{
		{
			name:       "TestNullable",
			joinColumn: "context_null_id",
			key:        "key1000",
			value:      "value1000",
		},
		{
			name:       "TestNotNullable",
			joinColumn: "context_id",
			key:        "key1000",
			value:      "value1000",
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			// Record the start time
			startTime := time.Now()

			// Begin a transaction
			tx, err := s.db.Begin()
			s.Require().Nil(err)

			// Prepare a statement for inserting data
			contextStmt, err := tx.Prepare("SELECT * FROM metrics LEFT JOIN contexts ON metrics." + tt.joinColumn + " = contexts.id WHERE contexts.json->>? = ?")
			s.Require().Nil(err)
			_, err = contextStmt.Exec(tt.key, tt.value) // Replace 'i' with the actual value you want to insert
			s.Require().Nil(err)

			defer contextStmt.Close() // Close the statement when we're done with it

			// Record the end time and calculate the duration
			endTime := time.Now()
			duration := endTime.Sub(startTime)

			// Print the duration
			s.T().Logf("Duration: %v", duration)
		})
	}
}
