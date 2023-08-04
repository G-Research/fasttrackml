package namespace

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
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

func GetFromFiber(c *fiber.Ctx) string {
	return GetFromContext(c.Context())
}

func GetFromContext(c context.Context) string {
	ns, ok := c.Value(key).(string)
	if !ok {
		return defaultValue
	}
	return ns
}
