package helpers

import (
	"encoding/json"
	"strings"
)

// ExtractContextBytes finds the metric context in decoded data, using the given key prefix,
// and marshals to bytes.
func ExtractContextBytes(contextPrefix string, decodedData map[string]any) ([]byte, error) {
	contx := ExtractContext(contextPrefix, decodedData)
	return json.Marshal(contx)
}

// ExtractContextBytes finds the metric context in the decoded data, using the given key prefix.
func ExtractContext(contextPrefix string, decodedData map[string]any) map[string]any {
	contx := map[string]any{}
	for key := range decodedData {
		if strings.HasPrefix(key, contextPrefix) && len(key) > len(contextPrefix) {
			contx[key[len(contextPrefix)+1:]] = decodedData[key]
		}
	}
	return contx
}
