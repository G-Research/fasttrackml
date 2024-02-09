package storage

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetArtifact_Ok(t *testing.T) {
	// setup
	runArtifactRoot := t.TempDir()
	fileName := "file.txt"
	fileContent := "artifact content"

	// #nosec G304
	f, err := os.Create(filepath.Join(runArtifactRoot, fileName))
	require.Nil(t, err)
	_, err = f.Write([]byte(fileContent))
	require.Nil(t, err)

	// invoke
	storage, err := NewLocal(nil)
	require.Nil(t, err)

	file, err := storage.Get(context.Background(), runArtifactRoot, fileName)
	require.Nil(t, err)
	//nolint:errcheck
	defer file.Close()

	// verify
	assert.NotNil(t, file)
	readBuffer := make([]byte, 20)
	ln, err := file.Read(readBuffer)
	require.Nil(t, err)
	assert.Equal(t, fileContent, string(readBuffer[:ln]))
}

func TestGetArtifact_Error(t *testing.T) {
	// setup
	runArtifactRoot := t.TempDir()
	subdir := "subdir"

	err := os.MkdirAll(filepath.Join(runArtifactRoot, subdir), os.ModePerm)
	require.Nil(t, err)

	// invoke
	storage, err := NewLocal(nil)
	require.Nil(t, err)

	file, err := storage.Get(context.Background(), runArtifactRoot, "non-existent-file")
	assert.NotNil(t, err)
	assert.Nil(t, file)

	// verify
	assert.Nil(t, file)
	assert.NotNil(t, err)

	// test subdir
	subdirFile, err := storage.Get(context.Background(), runArtifactRoot, subdir)
	assert.Nil(t, subdirFile)
	assert.NotNil(t, err)
}

func TestLocal_ListArtifacts_Ok(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runArtifactDir := t.TempDir()
			runArtifactURI := tt.prefix + runArtifactDir

			// 1. create test artifacts.
			err := os.WriteFile(filepath.Join(runArtifactDir, "artifact.file1"), []byte("contextX"), fs.ModePerm)
			require.Nil(t, err)
			err = os.Mkdir(filepath.Join(runArtifactDir, "artifact.dir"), fs.ModePerm)
			require.Nil(t, err)
			err = os.WriteFile(filepath.Join(runArtifactDir, "artifact.dir", "artifact.file2"), []byte("contentXX"), fs.ModePerm)
			require.Nil(t, err)

			// 2. create storage.
			storage, err := NewLocal(nil)
			require.Nil(t, err)

			// 3. list artifacts for root dir.
			rootDirResp, err := storage.List(context.Background(), runArtifactURI, "")
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
			require.Nil(t, err)

			// 4. list artifacts for sub dir.
			subDirResp, err := storage.List(context.Background(), runArtifactURI, "artifact.dir")
			assert.Equal(t, 1, len(subDirResp))
			assert.Equal(t, ArtifactObject{
				Path:  "artifact.dir/artifact.file2",
				IsDir: false,
				Size:  9,
			}, subDirResp[0])
			require.Nil(t, err)

			// 5. list artifacts for non-existing dir.
			nonExistingDirResp, err := storage.List(context.Background(), runArtifactURI, "non-existing-dir")
			assert.Equal(t, 0, len(nonExistingDirResp))
			require.Nil(t, err)
		})
	}
}
