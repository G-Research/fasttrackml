package mlflow

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	mlflow "github.com/G-Research/fasttrackml-ui-mlflow"
	"github.com/G-Research/fasttrackml/pkg/ui/common"
)

func AddRoutes(r fiber.Router) {
	r.Use("/static/mlflow/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(mlflow.FS),
	}))

	r.Use("/mlflow/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(
			common.NewOnlyRootFS(mlflow.FS, "index.html"),
		),
	}))
}
