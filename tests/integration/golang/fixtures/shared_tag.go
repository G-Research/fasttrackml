package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
)

// SharedTagFixtures represents data fixtures object.
type SharedTagFixtures struct {
	baseFixtures
}

// NewSharedTagFixtures creates new instance of TagFixtures.
func NewSharedTagFixtures(db *gorm.DB) (*SharedTagFixtures, error) {
	return &SharedTagFixtures{
		baseFixtures: baseFixtures{db: db},
	}, nil
}

// CreateTag creates new SharedTag.
func (f SharedTagFixtures) CreateTag(ctx context.Context, tagName string, namespaceID uint) (*models.SharedTag, error) {
	tag := models.SharedTag{
		Name:        tagName,
		NamespaceID: namespaceID,
	}
	if err := f.baseFixtures.db.WithContext(ctx).Create(&tag).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test tag")
	}
	return &tag, nil
}

// GetByRunID returns SharedTag list by requested Run ID.
func (f SharedTagFixtures) GetByRunID(ctx context.Context, runID string) ([]models.SharedTag, error) {
	var run models.Run
	if err := f.db.WithContext(ctx).Where(
		models.Run{ID: runID},
	).Preload("SharedTags").First(&run).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting tags by run id: %s", runID)
	}
	return run.SharedTags, nil
}

// Associate creates a tag relationship between run and tag.
func (f SharedTagFixtures) Associate(ctx context.Context, runID, tagID string) error {
	if err := f.db.WithContext(ctx).Exec(`
		INSERT INTO run_shared_tags (run_id, shared_tag_id) 
		VALUES (?, ?) 
		ON CONFLICT DO NOTHING`, runID, tagID).Error; err != nil {
		return eris.Wrapf(err, "error getting tags by run id: %s", runID)
	}
	return nil
}
