package helpers

import (
	"fmt"
	"strings"

	"golang.org/x/exp/rand"
)

// StrReplace makes replacing of multiple placeholders by theirs values in a string.
func StrReplace(str string, original []string, replacement []interface{}) string {
	for i, replace := range original {
		str = strings.NewReplacer(fmt.Sprintf("%v", replace), fmt.Sprintf("%v", replacement[i])).Replace(str)
	}

	return str
}

// GenerateRandomString generate random string with length equal n.
func GenerateRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
