package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
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
