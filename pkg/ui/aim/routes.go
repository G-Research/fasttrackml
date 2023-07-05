package aim

import (
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	aim "github.com/G-Research/fasttrackml-ui-aim"
)

type singleFileFS struct {
	fs.FS
	Path string
}

func (f singleFileFS) Open(name string) (fs.File, error) {
	return f.FS.Open(f.Path)
}

func AddRoutes(r fiber.Router) {
	r.Use("/static-files/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(aim.FS),
	}))

	r.Use("/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(singleFileFS{
			aim.FS,
			"index.html",
		}),
	}))
}
