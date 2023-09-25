package controller

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test_getFilename test the getFilename private helper.
func Test_getFilename(t *testing.T) {
	path := "/path/to/file/test.txt"
	expected := "test.txt"
	result := getFilename(path)
	assert.Equal(t, expected, result)
}

// Test_getContentType the Test_getContentType private helpers
func Test_getContentType(t *testing.T) {
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
		result := getContentType(testCase.filename)
		assert.Equal(t, testCase.expected, result, "Unexpected content type for filename: %s", testCase.filename)
	}
}
