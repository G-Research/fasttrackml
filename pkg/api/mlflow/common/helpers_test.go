package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContentType(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		result := GetContentType(tt.filename)
		assert.Equal(t, tt.expected, result, "Unexpected content type for filename: %s", tt.filename)
	}
}
