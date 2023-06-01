package helpers

import (
	"fmt"
	"strings"
)

// StrReplace makes replacing of multiple placeholders by theirs values in a string.
func StrReplace(str string, original []string, replacement []interface{}) string {
	for i, replace := range original {
		str = strings.NewReplacer(fmt.Sprintf("%v", replace), fmt.Sprintf("%v", replacement[i])).Replace(str)
	}

	return str
}
