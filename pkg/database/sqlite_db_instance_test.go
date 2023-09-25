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

	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := removeFile(tc.q, tc.dsnURL, tc.reset)
			if (err != nil) != tc.expectErr {
				t.Errorf("Test case '%s' failed. Got error: %v, expected error: %v", tc.name, err, tc.expectErr)
			}

			_, err = os.Stat(tempFile.Name()) // Check if the file still exists
			if tc.expectFileRemoved {
				assert.NotNil(t, err)
				assert.IsType(t, &fs.PathError{}, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_configureQuery(t *testing.T) {
	testCases := []struct {
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
				"mode":                 {"memory"},
			},
		},
	}

	for _, testCase := range testCases {
		result := configureQuery(testCase.inputURL)
		assert.Equal(t, testCase.expectedValues, result)
	}
}
