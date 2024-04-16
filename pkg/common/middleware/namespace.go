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
	return func(ctx *fiber.Ctx) (err error) {
		log.Debugf("checking namespace for path: %s", ctx.Path())
		// if namespace exists in the request then try to process it, otherwise fallback to default namespace.
		namespaceCode := models.DefaultNamespaceCode
		if matches := namespaceRegexp.FindStringSubmatch(ctx.Path()); matches != nil {
			namespaceCode = strings.Clone(matches[1])
			ctx.Path(strings.TrimPrefix(ctx.Path(), fmt.Sprintf("/ns/%s", namespaceCode)))
		}
		namespace, err := namespaceRepository.GetByCode(ctx.Context(), namespaceCode)
		if err != nil {
			return ctx.JSON(api.NewInternalError("error getting namespace with code: %s", namespaceCode))
		}
		if namespace == nil {
			return ctx.Status(
				http.StatusNotFound,
			).JSON(
				api.NewResourceDoesNotExistError("unable to find namespace with code: %s", namespaceCode),
			)
		}

		ctx.Locals(namespaceContextKey, namespace)

		return ctx.Next()
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
