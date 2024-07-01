package repositories

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/rotisserie/eris"
)

// ArtifactRepositoryProvider provides an interface to work with `artifact` entity.
type ArtifactRepositoryProvider interface {
	// Search will find artifacts based on the request.
	Search(
		ctx context.Context, namespaceID uint, timeZoneOffset int, req request.SearchArtifactsRequest,
	) (*sql.Rows, int64, SearchResultMap, error)
}

// ArtifactRepository repository to work with `artifact` entity.
type ArtifactRepository struct {
	db *gorm.DB
}

// NewArtifactRepository creates a repository to work with `artifact` entity.
func NewArtifactRepository(db *gorm.DB) *ArtifactRepository {
	return &ArtifactRepository{
		db: db,
	}
}

// Search will find artifacts based on the request.
func (r ArtifactRepository) Search(
	ctx context.Context, namespaceID uint, timeZoneOffset int, req request.SearchArtifactsRequest,
) (*sql.Rows, int64, SearchResultMap, error) {
	return nil, 0, nil, eris.New("Search function not implemented")
}
