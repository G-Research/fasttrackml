package namespace

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/admin/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
)

const (
	key         = "namespace"
	defaultCode = "default"
)

var nsUrl = regexp.MustCompile(`^/ns/([^/]+)/`)

func New(namespaceRepository repositories.NamespaceRepositoryProvider) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		log.Debugf("checking namespace for path: %s", c.Path())
		// if namespace exists in the request then try to process it, otherwise fallback to default namespace.
		if matches := nsUrl.FindStringSubmatch(c.Path()); matches != nil {
			namespaceCode := strings.Clone(matches[1])
			namespace, err := namespaceRepository.GetByCode(c.Context(), namespaceCode)
			if err != nil {
				return c.JSON(api.NewInternalError("error getting namespace with code: %s", namespaceCode))
			}
			if namespace == nil {
				return c.JSON(api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespaceCode))
			}

			c.Locals(key, namespace)
			c.Path(strings.TrimPrefix(c.Path(), fmt.Sprintf("/ns/%s", namespaceCode)))
		} else {
			namespace, err := namespaceRepository.GetByCode(c.Context(), defaultCode)
			if err != nil {
				return c.JSON(api.NewInternalError("error getting namespace with code: %s", defaultCode))
			}
			if namespace == nil {
				return c.JSON(api.NewResourceDoesNotExistError("unable to find namespace with code: %s", defaultCode))
			}
			c.Locals(key, namespace)
		}

		return c.Next()
	}
}

// GetNamespaceFromContext returns models.Namespace object from the context.
func GetNamespaceFromContext(ctx context.Context) (*models.Namespace, error) {
	namespace, ok := ctx.Value(key).(*models.Namespace)
	if !ok {
		return nil, eris.New("error getting namespace from context")
	}
	return namespace, nil
}
