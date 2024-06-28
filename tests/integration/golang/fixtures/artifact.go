package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
)

// ArtifactFixtures represents data fixtures object.
type ArtifactFixtures struct {
	baseFixtures
}

// NewArtifactFixtures creates new instance of ArtifactFixtures.
func NewArtifactFixtures(db *gorm.DB) (*ArtifactFixtures, error) {
	return &ArtifactFixtures{
		baseFixtures: baseFixtures{db: db},
	}, nil
}

// GetArtifactByRunID returns Run artifact by requested Run ID.
func (f ArtifactFixtures) GetArtifactByRunID(ctx context.Context, runID string) (*models.Artifact, error) {
	var artifact models.Artifact
	if err := f.db.WithContext(ctx).Where(
		models.Log{RunID: runID},
	).Find(&artifact).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting run artifact by run id: %s", runID)
	}
	return &artifact, nil
}
