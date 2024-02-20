package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/dto"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/query"
	"github.com/G-Research/fasttrackml/pkg/common/db/types"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// RunRepositoryProvider provides an interface to work with models.Run entity.
type RunRepositoryProvider interface {
	BaseRepositoryProvider
	// GetRunInfo returns run info.
	GetRunInfo(ctx context.Context, namespaceID uint, req *request.GetRunInfoRequest) (*models.Run, error)
	// GetRunMetrics returns Run metrics.
	GetRunMetrics(ctx context.Context, runID string, metricKeysMapDTO dto.MetricKeysMapDTO) ([]models.Metric, error)
	// GetAlignedMetrics returns aligned metrics.
	GetAlignedMetrics(
		ctx context.Context, namespaceID uint, values []any, alignBy string,
	) (*sql.Rows, func(*sql.Rows) (*models.AlignedMetric, error), error)
	// GetRunByNamespaceIDAndRunID returns experiment by Namespace ID and Run ID.
	GetRunByNamespaceIDAndRunID(ctx context.Context, namespaceID uint, runID string) (*models.Run, error)
	// GetByID returns models.Run entity by its ID.
	GetByID(ctx context.Context, id string) (*models.Run, error)
	// GetByNamespaceIDRunIDAndLifecycleStage returns models.Run entity by Namespace ID, its ID and Lifecycle Stage.
	GetByNamespaceIDRunIDAndLifecycleStage(
		ctx context.Context, namespaceID uint, runID string, lifecycleStage models.LifecycleStage,
	) (*models.Run, error)
	// GetByNamespaceIDAndStatus returns []models.Run by Namespace ID and status.
	GetByNamespaceIDAndStatus(ctx context.Context, namespaceID uint, status models.Status) ([]models.Run, error)
	// GetByNamespaceIDAndRunID returns models.Run entity by Namespace ID and its ID.
	GetByNamespaceIDAndRunID(
		ctx context.Context, namespaceID uint, runID string,
	) (*models.Run, error)
	// Create creates new models.Run entity.
	Create(ctx context.Context, run *models.Run) error
	// Update updates existing models.Experiment entity.
	Update(ctx context.Context, run *models.Run) error
	// Archive marks existing models.Run entity as archived.
	Archive(ctx context.Context, run *models.Run) error
	// Delete removes the existing models.Run
	Delete(ctx context.Context, namespaceID uint, run *models.Run) error
	// Restore marks existing models.Run entity as active.
	Restore(ctx context.Context, run *models.Run) error
	// ArchiveBatch marks existing models.Run entities as archived.
	ArchiveBatch(ctx context.Context, namespaceID uint, ids []string) error
	// DeleteBatch removes the existing models.Run from the db.
	DeleteBatch(ctx context.Context, namespaceID uint, ids []string) error
	// RestoreBatch marks existing models.Run entities as active.
	RestoreBatch(ctx context.Context, namespaceID uint, ids []string) error
	// SetRunTagsBatch sets Run tags in batch.
	SetRunTagsBatch(ctx context.Context, run *models.Run, batchSize int, tags []models.Tag) error
	// UpdateWithTransaction updates existing models.Run entity in scope of transaction.
	UpdateWithTransaction(ctx context.Context, tx *gorm.DB, run *models.Run) error
	// SearchRuns returns the list of runs by provided search request.
	SearchRuns(ctx context.Context, req request.SearchRunsRequest) ([]models.Run, int64, error)
}

// RunRepository repository to work with models.Run entity.
type RunRepository struct {
	BaseRepository
}

// NewRunRepository creates repository to work with models.Run entity.
func NewRunRepository(db *gorm.DB) *RunRepository {
	return &RunRepository{
		BaseRepository{
			db: db,
		},
	}
}

// GetRunInfo returns run info.
func (r RunRepository) GetRunInfo(
	ctx context.Context, namespaceID uint, req *request.GetRunInfoRequest,
) (*models.Run, error) {
	query := r.db.WithContext(ctx)
	for _, s := range req.Sequences {
		switch s {
		case "metric":
			query = query.Preload("LatestMetrics", func(db *gorm.DB) *gorm.DB {
				return db.Select("RunID", "Key", "ContextID")
			}).Preload(
				"LatestMetrics.Context",
			)
		}
	}

	run := models.Run{ID: req.ID}
	if err := query.InnerJoins(
		"Experiment",
		database.DB.Select(
			"ID", "Name",
		).Where(
			&models.Experiment{NamespaceID: namespaceID},
		),
	).Preload(
		"Params",
	).Preload(
		"Tags",
	).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting run info id: %s", req.ID)
	}
	return &run, nil
}

// GetRunMetrics returns Run metrics.
func (r RunRepository) GetRunMetrics(
	ctx context.Context, runID string, metricKeysMapDTO dto.MetricKeysMapDTO,
) ([]models.Metric, error) {
	subQuery := r.db.WithContext(ctx)
	for metricKey := range metricKeysMapDTO {
		subQuery = subQuery.Or("key = ? AND json = ?", metricKey.Name, types.JSONB(metricKey.Context))
	}

	// fetch run metrics based on provided criteria.
	var metrics []models.Metric
	if err := r.db.InnerJoins(
		"Context",
	).Order(
		"iter",
	).Where(
		"run_uuid = ?", runID,
	).Where(
		subQuery,
	).Find(&metrics).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting run metrics")
	}
	return metrics, nil
}

// GetAlignedMetrics returns aligned metrics.
func (r RunRepository) GetAlignedMetrics(
	ctx context.Context, namespaceID uint, values []any, alignBy string,
) (*sql.Rows, func(rows *sql.Rows) (*models.AlignedMetric, error), error) {
	var valuesStmt strings.Builder
	length := len(values) / 4
	for i := 0; i < length; i++ {
		valuesStmt.WriteString("(?, ?, CAST(? AS numeric), CAST(? AS numeric))")
		if i < length-1 {
			valuesStmt.WriteString(",")
		}
	}
	values = append(values, namespaceID, alignBy)
	rows, err := r.db.Raw(
		fmt.Sprintf("WITH params(run_uuid, key, context_id, steps) AS (VALUES %s)", &valuesStmt)+
			"        SELECT m.run_uuid, "+
			"				rm.key, "+
			"				m.iter, "+
			"				m.value, "+
			"				m.is_nan, "+
			"				rm.context_id, "+
			"				rm.context_json"+
			"		 FROM metrics AS m"+
			"        RIGHT JOIN ("+
			"          SELECT p.run_uuid, "+
			"				  p.key, "+
			"				  p.context_id, "+
			"				  lm.last_iter AS max, "+
			"				  (lm.last_iter + 1) / p.steps AS interval, "+
			"				  contexts.json AS context_json"+
			"          FROM params AS p"+
			"          LEFT JOIN latest_metrics AS lm USING(run_uuid, key, context_id)"+
			"          INNER JOIN contexts ON contexts.id = lm.context_id"+
			"        ) rm USING(run_uuid, context_id)"+
			"		 INNER JOIN runs AS r ON m.run_uuid = r.run_uuid"+
			"		 INNER JOIN experiments AS e ON r.experiment_id = e.experiment_id AND e.namespace_id = ?"+
			"        WHERE m.key = ?"+
			"          AND m.iter <= rm.max"+
			"          AND MOD(m.iter + 1 + rm.interval / 2, rm.interval) < 1"+
			"        ORDER BY r.row_num DESC, rm.key, rm.context_id, m.iter",
		values...,
	).Rows()
	if err != nil {
		return nil, nil, eris.Wrap(err, "error searching aligned run metrics")
	}
	if err := rows.Err(); err != nil {
		return nil, nil, eris.Wrap(err, "error getting query result")
	}
	return rows, func(rows *sql.Rows) (*models.AlignedMetric, error) {
		var metric models.AlignedMetric
		if err := r.db.ScanRows(rows, &metric); err != nil {
			return nil, eris.Wrap(err, "error getting aligned metric")
		}
		return &metric, nil
	}, nil
}

// GetRunByNamespaceIDAndRunID returns experiment by Namespace ID and Run ID.
func (r RunRepository) GetRunByNamespaceIDAndRunID(
	ctx context.Context, namespaceID uint, runID string,
) (*models.Run, error) {
	var run models.Run
	if err := r.db.WithContext(ctx).Select(
		"ID",
	).InnerJoins(
		"Experiment",
		database.DB.Select(
			"ID",
		).Where(
			&models.Experiment{NamespaceID: namespaceID},
		),
	).Where(
		"run_uuid = ?", runID,
	).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting run by id: %s", runID)
	}
	return &run, nil
}

// GetByID returns models.Run entity by its ID.
func (r RunRepository) GetByID(ctx context.Context, id string) (*models.Run, error) {
	run := models.Run{ID: id}
	if err := r.db.WithContext(
		ctx,
	).Preload(
		"LatestMetrics",
	).Preload(
		"Params",
	).Preload(
		"Tags",
	).First(&run).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting 'run' entity by id: %s", id)
	}
	return &run, nil
}

// GetByNamespaceIDRunIDAndLifecycleStage returns models.Run entity by Namespace ID, its ID and Lifecycle Stage..
func (r RunRepository) GetByNamespaceIDRunIDAndLifecycleStage(
	ctx context.Context, namespaceID uint, runID string, lifecycleStage models.LifecycleStage,
) (*models.Run, error) {
	run := models.Run{ID: runID}
	if err := r.db.WithContext(
		ctx,
	).Preload(
		"LatestMetrics",
	).Preload(
		"Params",
	).Preload(
		"Tags",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).Where(
		`runs.lifecycle_stage = ?`, lifecycleStage,
	).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting 'run' entity by id: %s", runID)
	}
	return &run, nil
}

// GetByNamespaceIDAndStatus returns []models.Run by Namespace ID and Lifecycle Stage.
func (r RunRepository) GetByNamespaceIDAndStatus(
	ctx context.Context, namespaceID uint, status models.Status,
) ([]models.Run, error) {
	var runs []models.Run
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		InnerJoins(
			"Experiment",
			database.DB.Select(
				"ID", "Name",
			).Where(
				&models.Experiment{NamespaceID: namespaceID},
			),
		).
		Preload("LatestMetrics.Context").
		Limit(50).
		Order("start_time DESC").
		Find(&runs).Error; err != nil {
		return nil, eris.Wrapf(err, "error retrieving runs by lifecycle stage")
	}
	return runs, nil
}

// GetByNamespaceIDAndRunID returns models.Run entity by Namespace ID and its ID.
func (r RunRepository) GetByNamespaceIDAndRunID(
	ctx context.Context, namespaceID uint, runID string,
) (*models.Run, error) {
	run := models.Run{ID: runID}
	if err := r.db.WithContext(
		ctx,
	).Preload(
		"LatestMetrics",
	).Preload(
		"Params",
	).Preload(
		"Tags",
	).Joins(
		"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
		namespaceID,
	).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, eris.Wrapf(err, "error getting 'run' entity by id: %s", runID)
	}
	return &run, nil
}

// Create creates new models.Run entity.
func (r RunRepository) Create(ctx context.Context, run *models.Run) error {
	// Lock need to calculate row_num
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("LOCK TABLE runs").Error; err != nil {
				return err
			}
		}
		return tx.Create(&run).Error
	}); err != nil {
		return eris.Wrap(err, "error creating new 'run' entity")
	}
	return nil
}

// Update updates existing models.Run entity.
func (r RunRepository) Update(ctx context.Context, run *models.Run) error {
	if err := r.db.WithContext(ctx).Model(&run).Updates(run).Error; err != nil {
		return eris.Wrapf(err, "error updating run with id: %s", run.ID)
	}
	return nil
}

// Archive marks existing models.Run entity as archived.
func (r RunRepository) Archive(ctx context.Context, run *models.Run) error {
	run.DeletedTime = sql.NullInt64{
		Int64: time.Now().UTC().UnixMilli(),
		Valid: true,
	}
	run.LifecycleStage = models.LifecycleStageDeleted
	if err := r.db.WithContext(ctx).Model(&run).Updates(run).Error; err != nil {
		return eris.Wrapf(err, "error updating existing run with id: %s", run.ID)
	}

	return nil
}

// ArchiveBatch marks existing models.Run entities as archived.
func (r RunRepository) ArchiveBatch(ctx context.Context, namespaceID uint, ids []string) error {
	if err := r.db.WithContext(
		ctx,
	).Model(
		models.Run{},
	).Where(
		"run_uuid IN (?)",
		r.db.Model(
			models.Run{},
		).Select(
			"run_uuid",
		).Joins(
			"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
			namespaceID,
		).Where(
			"run_uuid IN (?)", ids,
		),
	).Updates(models.Run{
		DeletedTime: sql.NullInt64{
			Int64: time.Now().UTC().UnixMilli(),
			Valid: true,
		},
		LifecycleStage: models.LifecycleStageDeleted,
	}).Error; err != nil {
		return eris.Wrapf(err, "error updating existing runs with ids: %s", ids)
	}

	return nil
}

// Delete removes the existing models.Run from the db.
func (r RunRepository) Delete(ctx context.Context, namespaceID uint, run *models.Run) error {
	return r.DeleteBatch(ctx, namespaceID, []string{run.ID})
}

// DeleteBatch removes existing models.Run from the db.
func (r RunRepository) DeleteBatch(ctx context.Context, namespaceID uint, ids []string) error {
	if err := r.db.Transaction(func(tx *gorm.DB) error {
		runs := make([]models.Run, 0, len(ids))
		if err := tx.Clauses(
			clause.Returning{Columns: []clause.Column{{Name: "row_num"}}},
		).Where(
			"run_uuid IN (?)",
			r.db.Model(
				models.Run{},
			).Select(
				"run_uuid",
			).Joins(
				"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
				namespaceID,
			).Where(
				"run_uuid IN (?)", ids,
			),
		).Delete(
			&runs,
		).Error; err != nil {
			return eris.Wrapf(err, "error deleting existing runs with ids: %s", ids)
		}

		// verify deletion
		// NOTE: tx.RowsAffected does not provide correct number of deleted, using the returning slice instead
		if len(runs) != len(ids) {
			return eris.New("count of deleted runs does not match length of ids input (invalid run ID?)")
		}

		// renumber the remainder
		if err := r.renumberRows(tx, getMinRowNum(runs)); err != nil {
			return eris.Wrapf(err, "error renumbering runs.row_num")
		}
		return nil
	}); err != nil {
		return eris.Wrapf(err, "error deleting runs")
	}

	return nil
}

// Restore marks existing models.Run entity as active.
func (r RunRepository) Restore(ctx context.Context, run *models.Run) error {
	// Use UpdateColumns so we can reset DeletedTime to null
	if err := r.db.WithContext(ctx).Model(&run).UpdateColumns(map[string]any{
		"DeletedTime":    sql.NullInt64{},
		"LifecycleStage": database.LifecycleStageActive,
	}).Error; err != nil {
		return eris.Wrapf(err, "error updating existing run with id: %s", run.ID)
	}

	return nil
}

// RestoreBatch marks existing models.Run entities as active.
func (r RunRepository) RestoreBatch(ctx context.Context, namespaceID uint, ids []string) error {
	if err := r.db.WithContext(
		ctx,
	).Where(
		"run_uuid IN (?)",
		r.db.Model(
			models.Run{},
		).Select(
			"run_uuid",
		).Joins(
			"INNER JOIN experiments ON experiments.experiment_id = runs.experiment_id AND experiments.namespace_id = ?",
			namespaceID,
		).Where(
			"run_uuid IN (?)", ids,
		),
	).Updates(models.Run{
		DeletedTime:    sql.NullInt64{},
		LifecycleStage: models.LifecycleStageActive,
	}).Error; err != nil {
		return eris.Wrapf(err, "error updating existing runs with ids: %s", ids)
	}

	return nil
}

// UpdateWithTransaction updates existing models.Run entity in scope of transaction.
func (r RunRepository) UpdateWithTransaction(ctx context.Context, tx *gorm.DB, run *models.Run) error {
	if err := tx.WithContext(ctx).Model(&run).Updates(run).Error; err != nil {
		return eris.Wrapf(err, "error updating existing run with id: %s", run.ID)
	}
	return nil
}

// SetRunTagsBatch sets Run tags in batch.
func (r RunRepository) SetRunTagsBatch(ctx context.Context, run *models.Run, batchSize int, tags []models.Tag) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tag := range tags {
			switch tag.Key {
			case "mlflow.user":
				run.UserID = tag.Value
				if err := r.UpdateWithTransaction(ctx, tx, run); err != nil {
					return eris.Wrap(err, "error updating run 'user_id' field")
				}
			case "mlflow.runName":
				run.Name = tag.Value
				if err := r.UpdateWithTransaction(ctx, tx, run); err != nil {
					return eris.Wrap(err, "error updating run 'name' field")
				}
			}
		}

		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).CreateInBatches(&tags, batchSize).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// SearchRuns returns the list of runs by provided search request.
func (r RunRepository) SearchRuns(ctx context.Context, req request.SearchRunsRequest) ([]models.Run, int64, error) {
	qp := query.QueryParser{
		Default: query.DefaultExpression{
			Contains:   "run.archived",
			Expression: "not run.archived",
		},
		Tables: map[string]string{
			"runs":        "runs",
			"experiments": "Experiment",
		},
		TzOffset:  req.TimeZoneOffset,
		Dialector: r.db.Dialector.Name(),
	}
	pq, err := qp.Parse(req.Query)
	if err != nil {
		return nil, 0, eris.Wrap(err, "problem parsing query")
	}

	var total int64
	if tx := r.db.WithContext(ctx).
		Model(&database.Run{}).
		Count(&total); tx.Error != nil {
		return nil, 0, eris.Wrap(tx.Error, "unable to count total runs")
	}

	log.Debugf("Total runs: %d", total)

	tx := r.db.WithContext(ctx).
		InnerJoins(
			"Experiment",
			database.DB.Select(
				"ID", "Name",
			).Where(
				&models.Experiment{NamespaceID: req.NamespaceID},
			),
		).
		Order("row_num DESC")

	if !req.ExcludeParams {
		tx.Preload("Params")
		tx.Preload("Tags")
	}

	if !req.ExcludeTraces {
		tx.Preload("LatestMetrics.Context")
	}

	if req.Limit > 0 {
		tx.Limit(req.Limit)
	}
	if req.Offset != "" {
		run := &database.Run{
			ID: req.Offset,
		}
		if err := database.DB.Select(
			"row_num",
		).First(
			&run,
		).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, eris.Wrapf(err, "unable to find search runs offset %q", req.Offset)
		}
		tx.Where("row_num < ?", run.RowNum)
	}
	var runs []models.Run
	if err := pq.Filter(tx).Find(&runs).Error; err != nil {
		return nil, 0, eris.Wrap(err, "error searching runs")
	}
	log.Debugf("found %d runs", len(runs))
	return runs, total, nil
}

// getMinRowNum will find the lowest row_num for the slice of runs
// or 0 for an empty slice
func getMinRowNum(runs []models.Run) models.RowNum {
	var minRowNum models.RowNum
	for _, run := range runs {
		if minRowNum == models.RowNum(0) || run.RowNum < minRowNum {
			minRowNum = run.RowNum
		}
	}
	return minRowNum
}

// renumberRows will update the runs.row_num field with the correct ordinal
func (r RunRepository) renumberRows(tx *gorm.DB, startWith models.RowNum) error {
	if startWith < models.RowNum(0) {
		return eris.New("attempting to renumber with less than 0 row number value")
	}

	if tx.Dialector.Name() == "postgres" {
		if err := tx.Exec("LOCK TABLE runs").Error; err != nil {
			return eris.Wrap(err, "unable to lock table")
		}
	}

	if err := tx.Exec(
		`UPDATE runs
	         SET row_num = rows.new_row_num
                 FROM (
                   SELECT run_uuid, ROW_NUMBER() OVER (ORDER BY start_time) + ? - 1 as new_row_num
                   FROM runs
                   WHERE runs.row_num >= ?
                 ) as rows
	         WHERE runs.run_uuid = rows.run_uuid`,
		int64(startWith),
		int64(startWith)).Error; err != nil {
		return eris.Wrap(err, "error updating runs.row_num")
	}
	return nil
}
