package config

import (
	"path/filepath"
	"testing"

	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"
)

func TestServiceConfig_Validate_Ok(t *testing.T) {
	testData := []struct {
		name           string
		providedConfig *ServiceConfig
		expectedConfig *ServiceConfig
	}{
		{
			name: "DefaultArtifactRootHasS3Prefix",
			providedConfig: &ServiceConfig{
				DefaultArtifactRoot: "s3://bucket_name",
			},
			expectedConfig: &ServiceConfig{
				DefaultArtifactRoot: "s3://bucket_name",
			},
		},
		{
			name: "DefaultArtifactRootHasFilePrefixAndIsRelative",
			providedConfig: &ServiceConfig{
				DefaultArtifactRoot: "file://path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				DefaultArtifactRoot: (func() string {
					path, err := filepath.Abs("path1/path2/path3")
					assert.Nil(t, err)
					return path
				})(),
			},
		},
		{
			name: "DefaultArtifactRootHasFilePrefixAndIsAbsolute",
			providedConfig: &ServiceConfig{
				DefaultArtifactRoot: "file:///path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				DefaultArtifactRoot: "/path1/path2/path3",
			},
		},
		{
			name: "DefaultArtifactRootHasEmptyPrefixAndIsAbsolute",
			providedConfig: &ServiceConfig{
				DefaultArtifactRoot: "/path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				DefaultArtifactRoot: "/path1/path2/path3",
			},
		},
		{
			name: "DefaultArtifactRootHasEmptyPrefixAndIsRelative",
			providedConfig: &ServiceConfig{
				DefaultArtifactRoot: "path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				DefaultArtifactRoot: (func() string {
					path, err := filepath.Abs("path1/path2/path3")
					assert.Nil(t, err)
					return path
				})(),
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(t, tt.providedConfig.Validate())
			assert.Equal(t, tt.providedConfig.DefaultArtifactRoot, tt.expectedConfig.DefaultArtifactRoot)
		})
	}
}

func TestServiceConfig_Validate_Error(t *testing.T) {
	testData := []struct {
		name   string
		error  error
		config *ServiceConfig
	}{
		{
			name: "DefaultArtifactRootHasIncorrectFormat",
			error: eris.New(
				`error validating service configuration: error parsing 'default-artifact-root' flag:parse "incorrect_format_of_schema://something": first path segment in URL cannot contain colon`,
			),
			config: &ServiceConfig{
				DefaultArtifactRoot: "incorrect_format_of_schema://something",
			},
		},
		{
			name:  "DefaultArtifactRootHasUnsupportedSchema",
			error: eris.New("error validating service configuration: unsupported schema of 'default-artifact-root' flag"),
			config: &ServiceConfig{
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
