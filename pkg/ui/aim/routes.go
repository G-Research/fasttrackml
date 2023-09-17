package aim

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	aim "github.com/G-Research/fasttrackml-ui-aim"
	"github.com/G-Research/fasttrackml/pkg/ui/common"
)

func AddRoutes(r fiber.Router) {
	r.Use("/static/aim/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(aim.FS),
	}))

	r.Use("/aim", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(
			common.NewSingleFileFS(aim.FS, "index.html")),
	}))
}
