package ui

import (
	"embed"
	"io/fs"
)

//go:embed chooser
var chooserFS embed.FS

//go:embed mlflow/build
var mlflowFS embed.FS

//go:embed aim/build
var aimFS embed.FS

var (
	ChooserFS fs.FS
	MlflowFS  fs.FS
	AimFS     fs.FS
)

func init() {
	ChooserFS, _ = fs.Sub(chooserFS, "chooser")
	MlflowFS, _ = fs.Sub(mlflowFS, "mlflow/build")
	AimFS, _ = fs.Sub(aimFS, "aim/build")
}
