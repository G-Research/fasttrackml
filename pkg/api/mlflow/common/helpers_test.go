package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetFileName test the getFilename private helper.
func TestGetFileName(t *testing.T) {
	path := "/path/to/file/test.txt"
	expected := "test.txt"
	result := GetFilename(path)
	assert.Equal(t, expected, result)
}

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
