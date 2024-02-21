package repositories

import (
	"context"

	"github.com/G-Research/fasttrackml/pkg/database"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
)

// TagRepositoryProvider provides an interface to work with models.Tag entity.
type TagRepositoryProvider interface {
	BaseRepositoryProvider
	// GetTagsByNamespace returns the list of tags.
	GetTagsByNamespace(ctx context.Context, namespaceID uint) ([]models.Tag, error)
	// CreateExperimentTag creates new models.ExperimentTag entity connected to models.Experiment.
	CreateExperimentTag(ctx context.Context, experimentTag *models.ExperimentTag) error
	// GetParamKeysByParameters returns list of tag keys by requested parameters.
	GetParamKeysByParameters(ctx context.Context, namespaceID uint, experiments []int) ([]string, error)
}

// TagRepository repository to work with models.Tag entity.
type TagRepository struct {
	BaseRepository
}

// NewTagRepository creates repository to work with models.Tag entity.
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{
		BaseRepository{
			db: db,
		},
	}
}

// CreateExperimentTag creates new models.ExperimentTag entity connected to models.Experiment.
func (r TagRepository) CreateExperimentTag(ctx context.Context, experimentTag *models.ExperimentTag) error {
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(experimentTag).Error; err != nil {
		return eris.Wrapf(err, "error creating tag for experiment with id: %d", experimentTag.ExperimentID)
	}
	return nil
}

// GetTagsByNamespace returns the list of tags.
// TODO fix stub implementation
func (r TagRepository) GetTagsByNamespace(ctx context.Context, namespaceID uint) ([]models.Tag, error) {
	var tags []models.Tag
	if err := r.db.WithContext(ctx).Find(&tags).Error; err != nil {
		return nil, err
	}
	return []models.Tag{}, nil
}

// GetParamKeysByParameters returns list of tag keys by requested parameters.
func (r TagRepository) GetParamKeysByParameters(
	ctx context.Context, namespaceID uint, experiments []int,
) ([]string, error) {
	// fetch and process tags.
	query := r.db.WithContext(ctx).Model(
		&database.Tag{},
	).Joins(
		"JOIN runs USING(run_uuid)",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).Where(
		"runs.lifecycle_stage = ?", database.LifecycleStageActive,
	)
	if len(experiments) != 0 {
		query.Where("experiments.experiment_id IN ?", experiments)
	}

	var keys []string
	if err := query.Pluck("Key", &keys).Error; err != nil {
		return nil, eris.Wrap(err, "error getting tag keys by parameters")
	}
	return keys, nil
}
