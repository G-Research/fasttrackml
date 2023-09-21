package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// TestGetArtifact for the local storage implementation.
func TestGetArtifact_Ok(t *testing.T) {
	// setup
	serviceArtifactRoot := "/tmp"
	runArtifactRoot := "/run-artifact-root/"
	fileName := "file.txt"

	err := os.MkdirAll(serviceArtifactRoot+runArtifactRoot, os.ModePerm)
	assert.Nil(t, err)

	f, err := os.Create(serviceArtifactRoot + runArtifactRoot + fileName)
	assert.Nil(t, err)
	_, err = f.Write([]byte("artifact here"))
	assert.Nil(t, err)

	// invoke
	svcConfig := &config.ServiceConfig{
		ArtifactRoot: serviceArtifactRoot,
	}
	svc, err := NewLocal(svcConfig)
	assert.Nil(t, err)

	file, err := svc.GetArtifact(runArtifactRoot, fileName)
	assert.Nil(t, err)
	defer func() {
		file.Close()
		os.Remove(f.Name())
		os.Remove(serviceArtifactRoot + runArtifactRoot)
	}()

	// verify
	assert.NotNil(t, file)
	p := make([]byte, 20)
	ln, err := file.Read(p)
	assert.Nil(t, err)

	assert.Equal(t, "artifact here", string(p[:ln]))
}

func TestGetArtifact_Error(t *testing.T) {
	// setup
	serviceArtifactRoot := "/tmp"
	runArtifactRoot := "/run-artifact-root/"
	fileName := "file.txt"

	err := os.MkdirAll(serviceArtifactRoot+runArtifactRoot, os.ModePerm)
	assert.Nil(t, err)

	f, err := os.Create(serviceArtifactRoot + runArtifactRoot + fileName)
	assert.Nil(t, err)
	_, err = f.Write([]byte("artifact here"))
	assert.Nil(t, err)

	// invoke
	svcConfig := &config.ServiceConfig{
		ArtifactRoot: serviceArtifactRoot,
	}
	svc, err := NewLocal(svcConfig)
	assert.Nil(t, err)

	file, err := svc.GetArtifact(runArtifactRoot, "some-other-item")
	assert.NotNil(t, err)
	defer func() {
		file.Close()
		os.Remove(f.Name())
		os.Remove(serviceArtifactRoot + runArtifactRoot)
	}()

	// verify
	assert.Nil(t, file)
	assert.NotNil(t, err)
}
