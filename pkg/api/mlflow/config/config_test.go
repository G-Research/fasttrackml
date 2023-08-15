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
			name: "ArtifactRootHasS3Prefix",
			providedConfig: &ServiceConfig{
				ArtifactRoot: "s3://bucket_name",
			},
			expectedConfig: &ServiceConfig{
				ArtifactRoot: "s3://bucket_name",
			},
		},
		{
			name: "ArtifactRootHasFilePrefixAndIsRelative",
			providedConfig: &ServiceConfig{
				ArtifactRoot: "file://path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				ArtifactRoot: (func() string {
					path, err := filepath.Abs("path1/path2/path3")
					assert.Nil(t, err)
					return path
				})(),
			},
		},
		{
			name: "ArtifactRootHasFilePrefixAndIsAbsolute",
			providedConfig: &ServiceConfig{
				ArtifactRoot: "file:///path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				ArtifactRoot: "/path1/path2/path3",
			},
		},
		{
			name: "ArtifactRootHasEmptyPrefixAndIsAbsolute",
			providedConfig: &ServiceConfig{
				ArtifactRoot: "/path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				ArtifactRoot: "/path1/path2/path3",
			},
		},
		{
			name: "ArtifactRootHasEmptyPrefixAndIsRelative",
			providedConfig: &ServiceConfig{
				ArtifactRoot: "path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				ArtifactRoot: (func() string {
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
			assert.Equal(t, tt.providedConfig.ArtifactRoot, tt.expectedConfig.ArtifactRoot)
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
			name: "ArtifactRootHasIncorrectFormat",
			error: eris.New(
				`error validating service configuration: error parsing 'artifact-root' flag: parse "incorrect_format_of_schema://something": first path segment in URL cannot contain colon`,
			),
			config: &ServiceConfig{
				ArtifactRoot: "incorrect_format_of_schema://something",
			},
		},
		{
			name:  "ArtifactRootHasUnsupportedSchema",
			error: eris.New(`error validating service configuration: unsupported schema of 'artifact-root' flag`),
			config: &ServiceConfig{
				ArtifactRoot: "unsupported://something",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.error.Error(), tt.config.Validate().Error())
		})
	}
}
