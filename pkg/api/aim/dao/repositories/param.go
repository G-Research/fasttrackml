package repositories

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
)

// ParamRepositoryProvider provides an interface to work with models.Param entity.
type ParamRepositoryProvider interface {
	// GetParamKeysByParameters returns list of param keys by requested parameters.
	GetParamKeysByParameters(ctx context.Context, namespaceID uint, experimentNames []string) ([]string, error)
}

// ParamRepository repository to work with models.Param entity.
type ParamRepository struct {
	repositories.BaseRepositoryProvider
}

// NewParamRepository creates repository to work with models.Param entity.
func NewParamRepository(db *gorm.DB) *ParamRepository {
	return &ParamRepository{
		repositories.NewBaseRepository(db),
	}
}

// GetParamKeysByParameters returns list of param keys by requested parameters.
func (r ParamRepository) GetParamKeysByParameters(
	ctx context.Context, namespaceID uint, experimentNames []string,
) ([]string, error) {
	query := r.GetDB().WithContext(ctx).Distinct().Model(
		&models.Param{},
	).Joins(
		"JOIN runs USING(run_uuid)",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).Where(
		"runs.lifecycle_stage = ?", models.LifecycleStageActive,
	)
	if len(experimentNames) != 0 {
		query = query.Where("experiments.name IN ?", experimentNames)
	}
	var keys []string
	if err := query.Pluck("Key", &keys).Error; err != nil {
		return nil, eris.Wrap(err, "error getting param keys by parameters")
	}
	return keys, nil
}
