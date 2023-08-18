package namespace

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/database"
)

const (
	key          = "namespace"
	defaultValue = "default"
)

var nsUrl = regexp.MustCompile(`^/ns/([^/]+)/`)

func New() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		log.Debugf("Checking namespace for path: %s", c.Path())
		if matches := nsUrl.FindStringSubmatch(c.Path()); matches != nil {
			ns := strings.Clone(matches[1])
			c.Locals("namespace", ns)
			c.Path(strings.TrimPrefix(c.Path(), fmt.Sprintf("/ns/%s", ns)))
			log.Debugf("Namespace: %s", c.Locals("namespace"))
		}

		return c.Next()
	}
}

func GetCodeFromContext(c context.Context) string {
	ns, ok := c.Value(key).(string)
	if !ok {
		return defaultValue
	}
	return ns
}

func GetCodeFromFiber(c *fiber.Ctx) string {
	return GetCodeFromContext(c.Context())
}

func GetIDFromCode(db *gorm.DB, code string) (uint, error) {
	var ns database.Namespace
	if err := db.
		Select("ID").
		Where("code = ?", code).
		First(&ns).
		Error; err != nil {
		return 0, err
	}
	return ns.ID, nil
}

func GetIDFromContext(db *gorm.DB, c context.Context) (uint, error) {
	code := GetCodeFromContext(c)
	return GetIDFromCode(db, code)
}

func GetIDFromFiber(db *gorm.DB, c *fiber.Ctx) (uint, error) {
	return GetIDFromContext(db, c.Context())
}
