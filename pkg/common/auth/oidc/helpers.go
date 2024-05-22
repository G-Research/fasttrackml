package oidc

import "github.com/rotisserie/eris"

// ConvertAndNormaliseRoles converts claim roles. normalise it, because `roles`
// could be represented as a one role(string) or array of roles(slice).
func ConvertAndNormaliseRoles(in interface{}) ([]string, error) {
	element, ok := in.(string)
	if ok {
		return []string{element}, nil
	}

	slice, ok := in.([]interface{})
	if !ok {
		return nil, eris.New("unsupported type of roles. should be string or []string")
	}

	out := make([]string, len(slice), cap(slice))
	for i, element := range slice {
		if convertedElement, ok := element.(string); ok {
			out[i] = convertedElement
		}
	}

	return out, nil
}
