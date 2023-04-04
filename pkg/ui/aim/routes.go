package aim

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed embed/build
var content embed.FS

type singleFileFS struct {
	fs.FS
	Path string
}

func (f singleFileFS) Open(name string) (fs.File, error) {
	return f.FS.Open(f.Path)
}

func AddRoutes(r fiber.Router) {
	sub, _ := fs.Sub(content, "embed/build")

	r.Use("/static-files/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))

	r.Use("/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(singleFileFS{
			sub,
			"index.html",
		}),
	}))
}
