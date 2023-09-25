package database

import (
	"io/fs"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_removeFile(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "test-db")
	assert.Nil(t, err)
	err = tempFile.Close()
	assert.Nil(t, err)

	tests := []struct {
		name              string
		q                 url.Values
		dsnURL            url.URL
		reset             bool
		expectErr         bool
		expectFileRemoved bool
	}{
		{
			name:      "Reset does nothing for memory DB",
			q:         url.Values{"mode": []string{"memory"}},
			dsnURL:    url.URL{},
			reset:     true,
			expectErr: false,
		},
		{
			name:              "Reset removes file for disk DB",
			q:                 url.Values{"mode": []string{"disk"}},
			dsnURL:            url.URL{Path: tempFile.Name()},
			reset:             true,
			expectErr:         false,
			expectFileRemoved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := removeFile(tt.q, tt.dsnURL, tt.reset)
			if (err != nil) != tt.expectErr {
				t.Errorf("Test case '%s' failed. Got error: %v, expected error: %v", tt.name, err, tt.expectErr)
			}

			_, err = os.Stat(tempFile.Name()) // Check if the file still exists
			if tt.expectFileRemoved {
				assert.NotNil(t, err)
				assert.IsType(t, &fs.PathError{}, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_configureQuery(t *testing.T) {
	tests := []struct {
		inputURL       url.URL
		expectedValues url.Values
	}{
		{
			inputURL: url.URL{
				RawQuery: "mode=disk",
			},
			expectedValues: url.Values{
				"_case_sensitive_like": {"true"},
				"_mutex":               {"no"},
				"_journal":             {"WAL"},
				"_query_only":          {"true"},
				"mode":                 {"disk"},
			},
		},
		{
			inputURL: url.URL{
				RawQuery: "mode=memory",
			},
			expectedValues: url.Values{
				"_case_sensitive_like": {"true"},
				"_mutex":               {"no"},
				"_query_only":          {"true"},
				"mode":                 {"memory"},
			},
		},
	}

	for _, tt := range tests {
		result := configureQuery(tt.inputURL)
		assert.Equal(t, tt.expectedValues, result)
	}
}

func Test_logDsnURL(t *testing.T) {
	// Define test cases as a slice of structs
	tests := []struct {
		name           string
		inputURL       string
		expectedResult string
	}{
		{
			name:           "No _key parameter",
			inputURL:       "https://example.com/db?user=user123",
			expectedResult: "https://example.com/db?user=user123",
		},
		{
			name:           "With _key parameter",
			inputURL:       "https://example.com/db?user=user123&_key=secret",
			expectedResult: "https://example.com/db?user=user123&_key=xxxxx",
		},
		// Add more test cases as needed
	}

	// Iterate through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input URL
			dsnURL, err := url.Parse(tt.inputURL)
			assert.NoError(t, err)

			// Call the function
			logDsnURL(dsnURL)

			// Parse the expected result URL
			expectedURL, err := url.Parse(tt.expectedResult)
			assert.NoError(t, err)

			// Assert the query values
			assert.Equal(t, expectedURL.Query(), dsnURL.Query())
		})
	}
}
