package repositories

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim/query"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
	"github.com/rotisserie/eris"
)

// ArtifactRepositoryProvider provides an interface to work with `artifact` entity.
type ArtifactRepositoryProvider interface {
	repositories.BaseRepositoryProvider
	// Search will find artifacts based on the request.
	Search(
		ctx context.Context,
		namespaceID uint,
		timeZoneOffset int,
		req request.SearchArtifactsRequest,
	) (*sql.Rows, int64, SearchResultMap, error)
}

// ArtifactRepository repository to work with `artifact` entity.
type ArtifactRepository struct {
	repositories.BaseRepositoryProvider
}

// NewArtifactRepository creates a repository to work with `artifact` entity.
func NewArtifactRepository(db *gorm.DB) *ArtifactRepository {
	return &ArtifactRepository{
		repositories.NewBaseRepository(db),
	}
}

// Search will find artifacts based on the request.
func (r ArtifactRepository) Search(
	ctx context.Context,
	namespaceID uint,
	timeZoneOffset int,
	req request.SearchArtifactsRequest,
) (*sql.Rows, int64, SearchResultMap, error) {
	qp := query.QueryParser{
		Default: query.DefaultExpression{
			Contains:   "run.archived",
			Expression: "not run.archived",
		},
		Tables: map[string]string{
			"runs":        "runs",
			"experiments": "experiments",
		},
		TzOffset:  timeZoneOffset,
		Dialector: r.GetDB().Dialector.Name(),
	}
	pq, err := qp.Parse(req.Query)
	if err != nil {
		return nil, 0, nil, err
	}

	var totalRuns int64
	if err := r.GetDB().WithContext(ctx).Model(&models.Run{}).Count(&totalRuns).Error; err != nil {
		return nil, 0, nil, eris.Wrap(err, "error counting metrics")
	}

	var runs []models.Run
	if tx := r.GetDB().WithContext(ctx).
		InnerJoins(
			"Experiment",
			r.GetDB().WithContext(ctx).Select(
				"ID", "Name",
			).Where(&models.Experiment{NamespaceID: namespaceID}),
		).
		Preload("Params").
		Preload("Tags").
		Where("run_uuid IN (?)", pq.Filter(r.GetDB().WithContext(ctx).
			Select("runs.run_uuid").
			Table("runs").
			Joins(
				"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
				namespaceID,
			),
		)).
		Order("runs.row_num DESC").
		Find(&runs); tx.Error != nil {
		return nil, 0, nil, eris.Wrap(err, "error searching artifacts")
	}

	runIDs := []string{}
	for _, run := range runs {
		runIDs = append(runIDs, run.ID)
	}
	result := make(SearchResultMap, len(runs))

	tx := r.GetDB().WithContext(ctx).
		Select(`row_number() over (order by run_uuid, step, created_at) as row_num, *`).
		Table("artifacts").
		Where("run_uuid IN ?", runIDs).
		Order("metrics.run_uuid").
		Order("metrics.step").
		Order("metrics.created_at")

	rows, err := tx.Rows()
	if err != nil {
		return nil, 0, nil, eris.Wrap(err, "error searching artifacts")
	}
	if err := rows.Err(); err != nil {
		return nil, 0, nil, eris.Wrap(err, "error getting artifacts rows cursor")
	}

	return rows, int64(len(runs)), result, nil
}
