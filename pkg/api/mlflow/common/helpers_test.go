package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetContentType the TestGetContentType private helpers
func TestGetContentType(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "FromListOfTextTypes",
			filename: "document.mlproject",
			expected: "text/plain",
		},
		{
			name:     "FromMimePackage",
			filename: "image.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "FromMimePackage",
			filename: "script.pdf",
			expected: "application/pdf",
		},
		{
			name:     "DefaultUnknownType",
			filename: "unknown.unknown",
			expected: "application/octet-stream",
		},
	}

	for _, testCase := range testCases {
		result := GetContentType(testCase.filename)
		assert.Equal(t, testCase.expected, result, "Unexpected content type for filename: %s", testCase.filename)
	}
}
