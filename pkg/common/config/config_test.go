package config

import (
	"path/filepath"
	"testing"

	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate_Ok(t *testing.T) {
	testData := []struct {
		name           string
		providedConfig *Config
		expectedConfig *Config
	}{
		{
			name: "DefaultArtifactRootHasS3Prefix",
			providedConfig: &Config{
				DefaultArtifactRoot: "s3://bucket_name",
			},
			expectedConfig: &Config{
				DefaultArtifactRoot: "s3://bucket_name",
			},
		},
		{
			name: "DefaultArtifactRootHasFilePrefixAndIsRelative",
			providedConfig: &Config{
				DefaultArtifactRoot: "file://path1/path2/path3",
			},
			expectedConfig: &Config{
				DefaultArtifactRoot: (func() string {
					path, err := filepath.Abs("path1/path2/path3")
					require.Nil(t, err)
					return "file://" + path
				})(),
			},
		},
		{
			name: "DefaultArtifactRootHasFilePrefixAndIsAbsolute",
			providedConfig: &Config{
				DefaultArtifactRoot: "file:///path1/path2/path3",
			},
			expectedConfig: &Config{
				DefaultArtifactRoot: "file:///path1/path2/path3",
			},
		},
		{
			name: "DefaultArtifactRootHasEmptyPrefixAndIsAbsolute",
			providedConfig: &Config{
				DefaultArtifactRoot: "/path1/path2/path3",
			},
			expectedConfig: &Config{
				DefaultArtifactRoot: "file:///path1/path2/path3",
			},
		},
		{
			name: "DefaultArtifactRootHasEmptyPrefixAndIsRelative",
			providedConfig: &Config{
				DefaultArtifactRoot: "path1/path2/path3",
			},
			expectedConfig: &Config{
				DefaultArtifactRoot: (func() string {
					path, err := filepath.Abs("path1/path2/path3")
					require.Nil(t, err)
					return "file://" + path
				})(),
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			require.Nil(t, tt.providedConfig.Validate())
			assert.Equal(t, tt.providedConfig.DefaultArtifactRoot, tt.expectedConfig.DefaultArtifactRoot)
		})
	}
}

func TestConfig_Validate_Error(t *testing.T) {
	testData := []struct {
		name   string
		error  error
		config *Config
	}{
		{
			name: "DefaultArtifactRootHasIncorrectFormat",
			error: eris.New(
				`error validating service configuration: error parsing 'default-artifact-root' flag: parse ` +
					`"incorrect_format_of_schema://something": first path segment in URL cannot contain colon`,
			),
			config: &Config{
				DefaultArtifactRoot: "incorrect_format_of_schema://something",
			},
		},
		{
			name: "DefaultArtifactRootHasUnsupportedSchema",
			error: eris.New(
				"error validating service configuration: unsupported schema of 'default-artifact-root' flag",
			),
			config: &Config{
				DefaultArtifactRoot: "unsupported://something",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.error.Error(), tt.config.Validate().Error())
		})
	}
}
