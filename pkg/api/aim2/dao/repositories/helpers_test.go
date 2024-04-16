package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

func Test_makeSqlPlaceholders(t *testing.T) {
	tests := []struct {
		numberInEachSet int
		numberOfSets    int
		expectedResult  string
	}{
		{numberInEachSet: 1, numberOfSets: 1, expectedResult: "(?)"},
		{numberInEachSet: 2, numberOfSets: 1, expectedResult: "(?,?)"},
		{numberInEachSet: 1, numberOfSets: 2, expectedResult: "(?),(?)"},
		{numberInEachSet: 2, numberOfSets: 2, expectedResult: "(?,?),(?,?)"},
	}

	for _, tt := range tests {
		result := makeSqlPlaceholders(tt.numberInEachSet, tt.numberOfSets)
		assert.Equal(t, tt.expectedResult, result)
	}
}

func Test_makeParamConflictPlaceholdersAndValues(t *testing.T) {
	tests := []struct {
		params               []models.Param
		expectedPlaceholders string
		expectedValues       []interface{}
	}{
		{
			params:               []models.Param{{Key: "key1", Value: "value1", RunID: "run1"}},
			expectedPlaceholders: "(?,?,?)",
			expectedValues:       []interface{}{"key1", "value1", "run1"},
		},
		{
			params: []models.Param{
				{Key: "key1", Value: "value1", RunID: "run1"},
				{Key: "key2", Value: "value2", RunID: "run2"},
			},
			expectedPlaceholders: "(?,?,?),(?,?,?)",
			expectedValues:       []interface{}{"key1", "value1", "run1", "key2", "value2", "run2"},
		},
	}

	for _, tt := range tests {
		placeholders, values := makeParamConflictPlaceholdersAndValues(tt.params)
		assert.Equal(t, tt.expectedPlaceholders, placeholders)
		assert.Equal(t, tt.expectedValues, values)
	}
}

func TestBuildJsonCondition(t *testing.T) {
	tests := []struct {
		name             string
		dialector        string
		jsonColumnName   string
		jsonPathValueMap map[string]string
		expectedSQL      string
		expectedArgs     []interface{}
	}{
		{
			name:           "Postgres",
			dialector:      database.PostgresDialectorName,
			jsonColumnName: "contexts.json",
			jsonPathValueMap: map[string]string{
				"key1":        "value1",
				"key2.nested": "value2",
			},
			expectedSQL:  "contexts.json#>>? = ? AND contexts.json#>>? = ?",
			expectedArgs: []interface{}{"{key1}", "value1", "{key2,nested}", "value2"},
		},
		{
			name:           "Sqlite",
			dialector:      database.SQLiteDialectorName,
			jsonColumnName: "contexts.json",
			jsonPathValueMap: map[string]string{
				"key1":        "value1",
				"key2.nested": "value2",
			},
			expectedSQL:  "contexts.json->>? = ? AND contexts.json->>? = ?",
			expectedArgs: []interface{}{"$.key1", "value1", "$.key2.nested", "value2"},
		},
		{
			name:             "SqliteEmptyMap",
			dialector:        database.SQLiteDialectorName,
			jsonColumnName:   "contexts.json",
			jsonPathValueMap: map[string]string{},
			expectedSQL:      "",
			expectedArgs:     []interface{}(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := BuildJsonCondition(tt.dialector, tt.jsonColumnName, tt.jsonPathValueMap)
			assert.Equal(t, tt.expectedSQL, sql)
			assert.ElementsMatch(t, tt.expectedArgs, args)
		})
	}
}
