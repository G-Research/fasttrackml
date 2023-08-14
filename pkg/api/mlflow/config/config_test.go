package config

import (
	"testing"

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
			name: "ArtifactRootHasFilePrefix",
			providedConfig: &ServiceConfig{
				ArtifactRoot: "file://path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				ArtifactRoot: "/pkg/api/mlflow/config/path1/path2/path3",
			},
		},
		{
			name: "ArtifactRootHasEmptyPrefix",
			providedConfig: &ServiceConfig{
				ArtifactRoot: "/path1/path2/path3",
			},
			expectedConfig: &ServiceConfig{
				ArtifactRoot: "/path1/path2/path3",
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			assert.Nil(t, tt.providedConfig.Validate())
			assert.Contains(t, tt.providedConfig.ArtifactRoot, tt.expectedConfig.ArtifactRoot)
		})
	}
}
