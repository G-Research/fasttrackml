package repositories

import (
	"fmt"
	"strings"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// makeSqlPlaceholders collects a string of (?,?,?), (?,?,?), etc
func makeSqlPlaceholders(numberInEachSet, numberOfSets int) string {
	placeholderArray := make([]string, numberInEachSet)
	for i := 0; i < numberInEachSet; i++ {
		placeholderArray[i] = "?"
	}
	setsArray := make([]string, numberOfSets)
	for i := 0; i < numberOfSets; i++ {
		setsArray[i] = fmt.Sprintf("(%s)", strings.Join(placeholderArray, ","))
	}
	return strings.Join(setsArray, ",")
}

// makeParamSqlValues concatenates Key, Value, RunID from each input Param for use in sql values replacement
func makeParamSqlValues(params []models.Param) []interface{} {
	// values array is params * 3 in length since using 3 fields from each
	valuesArray := make([]interface{}, len(params)*3)
	index := 0
	for _, param := range params {
		valuesArray[index] = param.Key
		valuesArray[index+1] = param.Value
		valuesArray[index+2] = param.RunID
		index = index + 3
	}
	return valuesArray
}
