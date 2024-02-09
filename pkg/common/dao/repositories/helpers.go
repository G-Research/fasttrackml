package repositories

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
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

// BuildJsonCondition creates sql and values for where condition to select items having the specified map of json paths
// and values in the given json column. Json path is expressed as "key" or "outerkey.nestedKey".
func BuildJsonCondition(
	dialector string,
	jsonColumnName string,
	jsonPathValueMap map[string]string,
) (sql string, args []any) {
	if len(jsonPathValueMap) == 0 {
		return sql, args
	}
	var conditionTemplate string
	args = make([]any, len(jsonPathValueMap)*2)
	switch dialector {
	case postgres.Dialector{}.Name():
		conditionTemplate = "%s#>>? = ?"
		idx := 0
		for k, v := range jsonPathValueMap {
			path := strings.ReplaceAll(k, ".", ",")
			args[idx] = fmt.Sprintf("{%s}", path)
			args[idx+1] = v
			idx = idx + 2
		}
	default:
		conditionTemplate = "%s->>? = ?"
		idx := 0
		for k, v := range jsonPathValueMap {
			args[idx] = fmt.Sprintf("$.%s", k)
			args[idx+1] = v
			idx = idx + 2
		}
	}
	conditionTemplate = fmt.Sprintf(conditionTemplate, jsonColumnName)
	sql = strings.Repeat(conditionTemplate+" AND ", len(jsonPathValueMap)-1) + conditionTemplate
	return sql, args
}
