package storage

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetArtifact for the local storage implementation.
func TestGetArtifact_Ok(t *testing.T) {
	// setup
	runArtifactRoot := t.TempDir()
	fileName := "file.txt"
	fileContent := "artifact content"

	f, err := os.Create(filepath.Join(runArtifactRoot, fileName))
	assert.Nil(t, err)
	_, err = f.Write([]byte(fileContent))
	assert.Nil(t, err)

	// invoke
	storage, err := NewLocal(nil)
	assert.Nil(t, err)

	file, err := storage.Get(runArtifactRoot, fileName)
	assert.Nil(t, err)
	defer file.Close()

	// verify
	assert.NotNil(t, file)
	readBuffer := make([]byte, 20)
	ln, err := file.Read(readBuffer)
	assert.Nil(t, err)
	assert.Equal(t, fileContent, string(readBuffer[:ln]))
}

func TestGetArtifact_Error(t *testing.T) {
	// setup
	runArtifactRoot := t.TempDir()
	subdir := "subdir"

	err := os.MkdirAll(filepath.Join(runArtifactRoot, subdir), os.ModePerm)
	assert.Nil(t, err)

	// invoke
	storage, err := NewLocal(nil)
	assert.Nil(t, err)

	file, err := storage.Get(filepath.Join(runArtifactRoot), "non-existent-file")
	assert.NotNil(t, err)
	assert.Nil(t, file)

	// verify
	assert.Nil(t, file)
	assert.NotNil(t, err)

	// test subdir
	subdirFile, err := storage.Get(filepath.Join(runArtifactRoot), subdir)
	assert.Nil(t, subdirFile)
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
