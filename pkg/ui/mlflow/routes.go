package mlflow

import (
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	mlflow "github.com/G-Research/fasttrackml-ui-mlflow"
)

type onlyRootFS struct {
	fs.FS
	Path string
}

func (f onlyRootFS) Open(name string) (fs.File, error) {
	if name != "." {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return f.FS.Open(f.Path)
}

func AddRoutes(r fiber.Router) {
	r.Use("/static-files/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(mlflow.FS),
	}))

	r.Use("/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(onlyRootFS{
			mlflow.FS,
			"index.html",
		}),
	}))
}
