package common

import "io/fs"

type singleFileFS struct {
	fs.FS
	path string
}

func NewSingleFileFS(fs fs.FS, path string) fs.FS {
	return &singleFileFS{
		fs,
		path,
	}
}

func (f *singleFileFS) Open(name string) (fs.File, error) {
	return f.FS.Open(f.path)
}

type onlyRootFS struct {
	fs.FS
	path string
}

func NewOnlyRootFS(fs fs.FS, path string) fs.FS {
	return &onlyRootFS{
		fs,
		path,
	}
}

func (f onlyRootFS) Open(name string) (fs.File, error) {
	if name != "." {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return f.FS.Open(f.path)
}
