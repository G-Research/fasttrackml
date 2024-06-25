package convertors

import (
	"github.com/google/uuid"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
)

// ConvertCreateTagRequestToDBModel translates the request to a model.
func ConvertCreateTagRequestToDBModel(req request.CreateTagRequest, namespaceID uint) models.SharedTag {
	return models.SharedTag{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		NamespaceID: namespaceID,
	}
}
