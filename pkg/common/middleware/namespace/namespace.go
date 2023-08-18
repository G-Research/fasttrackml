package namespace

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/G-Research/fasttrackml/pkg/common/dao/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"

	"github.com/G-Research/fasttrackml/pkg/api/admin/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
)

const (
	key          = "namespace"
	defaultValue = "default"
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

			c.Locals("namespace", namespace)
			c.Path(strings.TrimPrefix(c.Path(), fmt.Sprintf("/ns/%s", namespaceCode)))
			log.Debugf("namespace: %s", c.Locals("namespace"))
		} else {
			namespace, err := namespaceRepository.GetByCode(c.Context(), defaultValue)
			if err != nil {
				return c.JSON(api.NewInternalError("error getting namespace with code: %s", defaultValue))
			}
			if namespace == nil {
				return c.JSON(api.NewResourceDoesNotExistError("unable to find namespace with code: %s", defaultValue))
			}

			c.Locals("namespace", namespace)
		}

		return c.Next()
	}
}

func GetNamespaceFromContext(ctx context.Context) (*models.Namespace, error) {
	namespace, ok := ctx.Value(key).(*models.Namespace)
	if !ok {
		return nil, eris.New("error getting namespace from context")
	}
	return namespace, nil
}
