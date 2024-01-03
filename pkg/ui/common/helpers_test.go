package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorMessageForUI(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		errMsg   string
		expected string
	}{
		{
			name:     "UniqueError",
			field:    "email",
			errMsg:   "UNIQUE CONSTRAINT: Duplicate entry 'test@example.com' for key 'email'",
			expected: "The email is already in use.",
		},
		{
			name:     "ValidationError",
			field:    "password",
			errMsg:   "INVALID_PARAMETER_VALUE: Password must be at least 8 characters long",
			expected: "The password is invalid.",
		},
		{
			name:     "UnknownError",
			field:    "username",
			errMsg:   "An unknown error occurred",
			expected: "An unexpected error was encountered: An unknown error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ErrorMessageForUI(tt.field, tt.errMsg)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
