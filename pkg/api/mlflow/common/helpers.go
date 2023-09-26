package common

import (
	"mime"
	"path"
	"strings"

	"golang.org/x/exp/slices"
)

// textTypes used by GetContentType.
var textTypes = []string{
	"txt",
	"log",
	"err",
	"cfg",
	"conf",
	"cnf",
	"cf",
	"ini",
	"properties",
	"prop",
	"hocon",
	"toml",
	"yaml",
	"yml",
	"xml",
	"json",
	"js",
	"py",
	"py3",
	"csv",
	"tsv",
	"md",
	"rst",
	"MLmodel",
	"mlproject",
}

// GetPointer returns pointer for provided string.
func GetPointer[T any](str T) *T {
	return &str
}

// GetFilename returns the final bit of the path (the filename).
func GetFilename(fullpath string) string {
	_, filename := path.Split(fullpath)
	return filename
}

// GetContentType will determine the content type of the file.
func GetContentType(filename string) string {
	fileExt := path.Ext(filename)
	if slices.Contains(textTypes, strings.Trim(fileExt, ".")) {
		return "text/plain"
	}
	mimeType := mime.TypeByExtension(fileExt)
	if mimeType != "" {
		return mimeType
	}
	return "application/octet-stream"
}
