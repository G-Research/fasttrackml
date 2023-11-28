package repositories

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
)

// makeSqlPlaceholders collects a string of "(?,?,?), (?,?,?)" and so on,
// for use as sql parameters
func makeSqlPlaceholders(numberInEachSet, numberOfSets int) string {
	set := fmt.Sprintf("(%s)", strings.Repeat("?,", numberInEachSet-1)+"?")
	return strings.Repeat(set+",", numberOfSets-1) + set
}

// makeParamConflictPlaceholdersAndValues provides sql placeholders and concatenates
// Key, Value, RunID from each input Param for use in sql values replacement
func makeParamConflictPlaceholdersAndValues(params []models.Param) (string, []interface{}) {
	// make place holders of 3 fields for each param
	placeholders := makeSqlPlaceholders(3, len(params))
	// values array is params * 3 in length since using 3 fields from each
	valuesArray := make([]interface{}, len(params)*3)
	index := 0
	for _, param := range params {
		valuesArray[index] = param.Key
		valuesArray[index+1] = param.Value
		valuesArray[index+2] = param.RunID
		index = index + 3
	}
	return placeholders, valuesArray
}

// AddJsonCondition adds where condition(s) to the query to select items having the specified map of json paths
// and values in the given json column. Json path is expressed as "key" or "outerkey.nestedKey".
func AddJsonCondition(tx *gorm.DB, jsonColumnName string, jsonPathValueMap map[string]string) {
	if len(jsonPathValueMap) == 0 {
		return
	}
	switch tx.Dialector.Name() {
	case postgres.Dialector{}.Name():
		for k, v := range jsonPathValueMap {
			path := strings.ReplaceAll(k, ".", ",")
			tx.Where(fmt.Sprintf("%s#>>'{%s}' = ?", jsonColumnName, path), v)
		}
	default:
		for k, v := range jsonPathValueMap {
			tx.Where(fmt.Sprintf("%s->>'%s' = ?", jsonColumnName, k), v)
		}
	}
}
