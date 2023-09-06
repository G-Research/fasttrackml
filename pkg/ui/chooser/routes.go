package chooser

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2/middleware/filesystem"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
)

//go:embed embed/*
var content embed.FS

func AddRoutes(r fiber.Router) {
	sub, _ := fs.Sub(content, "embed")

	// default route
	r.Use("/", etag.New(), filesystem.New(filesystem.Config{
		Root: http.FS(sub),
	}))
}
