package storage

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
)

// TestGetArtifact for the local storage implementation.
func TestGetArtifact_Ok(t *testing.T) {
	// setup
	defaultArtifactRoot := t.TempDir()
	runArtifactRoot := "/run-artifact-root/"
	fileName := "file.txt"
	fileContent := "artifact content"

	err := os.MkdirAll(filepath.Join(defaultArtifactRoot, runArtifactRoot), os.ModePerm)
	assert.Nil(t, err)

	f, err := os.Create(filepath.Join(defaultArtifactRoot, runArtifactRoot, fileName))
	assert.Nil(t, err)
	_, err = f.Write([]byte(fileContent))
	assert.Nil(t, err)

	// invoke
	svcConfig := &config.ServiceConfig{
		DefaultArtifactRoot: defaultArtifactRoot,
	}
	svc, err := NewLocal(svcConfig)
	assert.Nil(t, err)

	file, err := svc.GetArtifact(filepath.Join(defaultArtifactRoot, runArtifactRoot), fileName)
	assert.Nil(t, err)
	defer func() {
		file.Close()
		os.Remove(f.Name())
	}()

	// verify
	assert.NotNil(t, file)
	p := make([]byte, 20)
	ln, err := file.Read(p)
	assert.Nil(t, err)

	assert.Equal(t, fileContent, string(p[:ln]))
}

func TestGetArtifact_Error(t *testing.T) {
	// setup
	defaultArtifactRoot := t.TempDir()
	runArtifactRoot := "/run-artifact-root/"
	fileName := "file.txt"
	fileContent := "artifact content"

	err := os.MkdirAll(filepath.Join(defaultArtifactRoot, runArtifactRoot), os.ModePerm)
	assert.Nil(t, err)

	f, err := os.Create(filepath.Join(defaultArtifactRoot, runArtifactRoot, fileName))
	assert.Nil(t, err)
	_, err = f.Write([]byte(fileContent))
	assert.Nil(t, err)

	// invoke
	svcConfig := &config.ServiceConfig{
		DefaultArtifactRoot: defaultArtifactRoot,
	}
	svc, err := NewLocal(svcConfig)
	assert.Nil(t, err)

	file, err := svc.GetArtifact(filepath.Join(defaultArtifactRoot, runArtifactRoot), "some-other-item")
	assert.NotNil(t, err)
	defer func() {
		file.Close()
		os.Remove(f.Name())
	}()

	// verify
	assert.Nil(t, file)
	assert.NotNil(t, err)
}

func TestLocal_ListArtifacts_Ok(t *testing.T) {
	testData := []struct {
		name   string
		prefix string
	}{
		{
			name:   "TestWithFilePrefix",
			prefix: "file://",
		},
		{
			name:   "TestWithoutPrefix",
			prefix: "",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			runArtifactDir := t.TempDir()
			runArtifactURI := tt.prefix + runArtifactDir

			// 1. create test artifacts.
			err := os.WriteFile(filepath.Join(runArtifactDir, "artifact.file1"), []byte("contextX"), fs.ModePerm)
			assert.Nil(t, err)
			err = os.Mkdir(filepath.Join(runArtifactDir, "artifact.dir"), fs.ModePerm)
			assert.Nil(t, err)
			err = os.WriteFile(filepath.Join(runArtifactDir, "artifact.dir", "artifact.file2"), []byte("contentXX"), fs.ModePerm)
			assert.Nil(t, err)

			// 2. create storage.
			storage, err := NewLocal(nil)
			assert.Nil(t, err)

			// 3. list artifacts for root dir.
			rootDirResp, err := storage.List(runArtifactURI, "")
			assert.Equal(t, 2, len(rootDirResp))
			assert.Equal(t, []ArtifactObject{
				{
					Path:  "artifact.dir",
					IsDir: true,
					Size:  0,
				},
				{
					Path:  "artifact.file1",
					IsDir: false,
					Size:  8,
				},
			}, rootDirResp)
			assert.Nil(t, err)

			// 4. list artifacts for sub dir.
			subDirResp, err := storage.List(runArtifactURI, "artifact.dir")
			assert.Equal(t, 1, len(subDirResp))
			assert.Equal(t, ArtifactObject{
				Path:  "artifact.dir/artifact.file2",
				IsDir: false,
				Size:  9,
			}, subDirResp[0])
			assert.Nil(t, err)

			// 5. list artifacts for non-existing dir.
			nonExistingDirResp, err := storage.List(runArtifactURI, "non-existing-dir")
			assert.Equal(t, 0, len(nonExistingDirResp))
			assert.Nil(t, err)
		})
	}
}
