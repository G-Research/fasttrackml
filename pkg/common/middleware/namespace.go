package middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/api"
)

const (
	namespaceContextKey = "namespace"
)

var namespaceRegexp = regexp.MustCompile(`^/ns/([^/]+)/`)

// NewNamespaceMiddleware creates new Middleware instance.
func NewNamespaceMiddleware(namespaceRepository repositories.NamespaceRepositoryProvider) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		log.Debugf("checking namespace for path: %s", c.Path())
		// if namespace exists in the request then try to process it, otherwise fallback to default namespace.
		namespaceCode := models.DefaultNamespaceCode
		if matches := namespaceRegexp.FindStringSubmatch(c.Path()); matches != nil {
			namespaceCode = strings.Clone(matches[1])
			c.Path(strings.TrimPrefix(c.Path(), fmt.Sprintf("/ns/%s", namespaceCode)))
		}
		namespace, err := namespaceRepository.GetByCode(c.Context(), namespaceCode)
		if err != nil {
			return c.JSON(api.NewInternalError("error getting namespace with code: %s", namespaceCode))
		}
		if namespace == nil {
			return c.Status(
				http.StatusNotFound,
			).JSON(
				api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespaceCode),
			)
		}

		c.Locals(namespaceContextKey, namespace)

		return c.Next()
	}
}

// GetNamespaceFromContext returns models.Namespace object from the context.
func GetNamespaceFromContext(ctx context.Context) (*models.Namespace, error) {
	namespace, ok := ctx.Value(namespaceContextKey).(*models.Namespace)
	if !ok {
		return nil, eris.New("error getting namespace from context")
	}
	return namespace, nil
}
