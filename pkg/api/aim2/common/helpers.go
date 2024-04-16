package common

import (
	"encoding/json"
	"reflect"
)

// CompareJson compares two json objects.
func CompareJson(json1, json2 []byte) bool {
	var j, j2 interface{}
	if err := json.Unmarshal(json1, &j); err != nil {
		return false
	}
	if err := json.Unmarshal(json2, &j2); err != nil {
		return false
	}
	return reflect.DeepEqual(j2, j)
}
