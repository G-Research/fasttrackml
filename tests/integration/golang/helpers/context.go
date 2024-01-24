package helpers

import (
	"encoding/json"
	"strings"
)

// ExtractContextBytes will find the context from the decoded data, using the given key prefix.
func ExtractContextBytes(contextPrefix string, decodedData map[string]any) ([]byte, error) {
	contx := ExtractContext(contextPrefix, decodedData)
	return json.Marshal(contx)
}

// ExtractContextBytes will find the context from the decoded data, using the given key prefix,
// and marshal to bytes.
func ExtractContext(contextPrefix string, decodedData map[string]any) map[string]any {
	contx := map[string]any{}
	for key := range decodedData {
		if strings.HasPrefix(key, contextPrefix) && len(key) > len(contextPrefix) {
			contx[key[len(contextPrefix)+1:]] = decodedData[key]
		}
	}
	return contx
}
