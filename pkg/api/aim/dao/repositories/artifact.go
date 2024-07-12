package repositories

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/aim/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim/query"
	"github.com/G-Research/fasttrackml/pkg/common/dao/repositories"
)

// ArtifactSearchStepInfo is a search summary for a Run Step.
type ArtifactSearchStepInfo struct {
	RunUUID  string `gorm:"column:run_uuid"`
	Step     int    `gorm:"column:step"`
	ImgCount int    `gorm:"column:img_count"`
}

// ArtifactSearchSummary is a search summary for whole run.
type ArtifactSearchSummary map[string][]ArtifactSearchStepInfo

// TotalSteps figures out how many steps belong to the runID.
func (r ArtifactSearchSummary) TotalSteps(runID string) int {
	return len(r[runID])
}

// StepImageCount figures out how many steps belong to the runID and step.
func (r ArtifactSearchSummary) StepImageCount(runID string, step int) int {
	runStepImages := r[runID]
	return runStepImages[step].ImgCount
}

// ArtifactRepositoryProvider provides an interface to work with `artifact` entity.
type ArtifactRepositoryProvider interface {
	repositories.BaseRepositoryProvider
	// Search will find artifacts based on the request.
	Search(
		ctx context.Context,
		namespaceID uint,
		timeZoneOffset int,
		req request.SearchArtifactsRequest,
	) (*sql.Rows, map[string]models.Run, ArtifactSearchSummary, error)
	GetArtifactNamesByExperiments(
		ctx context.Context, namespaceID uint, experiments []int,
	) ([]string, error)
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
) (*sql.Rows, map[string]models.Run, ArtifactSearchSummary, error) {
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
		return nil, nil, nil, err
	}

	runIDs := []string{}
	runs := []models.Run{}
	if tx := pq.Filter(r.GetDB().WithContext(ctx).
		Table("runs").
		Joins(`INNER JOIN experiments
                        ON experiments.experiment_id = runs.experiment_id
                        AND experiments.namespace_id = ?`,
			namespaceID,
		)).
		Find(&runs); tx.Error != nil {
		return nil, nil, nil, eris.Wrap(err, "error finding runs for artifact search")
	}

	runMap := make(map[string]models.Run, len(runs))
	for _, run := range runs {
		runIDs = append(runIDs, run.ID)
		runMap[run.ID] = run
	}

	// collect some summary data for progress indicator
	stepInfo := []ArtifactSearchStepInfo{}
	if tx := r.GetDB().WithContext(ctx).
		Raw(`SELECT run_uuid, step, count(id) as img_count
			  FROM artifacts
			  WHERE run_uuid IN (?)
			  GROUP BY run_uuid, step;`, runIDs).
		Find(&stepInfo); tx.Error != nil {
		return nil, nil, nil, eris.Wrap(err, "error find result summary for artifact search")
	}

	resultSummary := make(ArtifactSearchSummary, len(runIDs))
	for _, rslt := range stepInfo {
		resultSummary[rslt.RunUUID] = append(resultSummary[rslt.RunUUID], rslt)
	}

	// get a cursor for the artifacts
	tx := r.GetDB().WithContext(ctx).
		Table("artifacts").
		Where("run_uuid IN ?", runIDs).
		Order("run_uuid").
		Order("step").
		Order("created_at")

	rows, err := tx.Rows()
	if err != nil {
		return nil, nil, nil, eris.Wrap(err, "error searching artifacts")
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, eris.Wrap(err, "error getting artifacts rows cursor")
	}

	return rows, runMap, resultSummary, nil
}

// GetArtifactNamesByExperiments will find image names in the selected experiments.
func (r ArtifactRepository) GetArtifactNamesByExperiments(
	ctx context.Context, namespaceID uint, experiments []int,
) ([]string, error) {
	runIDs := []string{}
	if err := r.GetDB().WithContext(ctx).
		Table("runs").
		Joins(`INNER JOIN experiments
                        ON experiments.experiment_id = runs.experiment_id
                        AND experiments.namespace_id = ?
		        AND experiments.experiment_id IN ?`,
			namespaceID, experiments,
		).
		Find(&runIDs).Error; err != nil {
		return nil, eris.Wrap(err, "error finding runs for artifacts")
	}

	imageNames := []string{}
	if err := r.GetDB().WithContext(ctx).
		Distinct("name").
		Table("artifacts").
		Where("run_uuid IN ?", runIDs).
		Find(&imageNames).Error; err != nil {
		return nil, eris.Wrap(err, "error finding runs for artifact search")
	}
	return imageNames, nil
}
