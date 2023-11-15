package aim

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
)

func GetExperiments(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperiments namespace: %s", ns.Code)

	var experiments []struct {
		database.Experiment
		RunCount    int
		Description string `gorm:"column:description"`
	}
	if tx := database.DB.Model(&database.Experiment{}).
		Select(
			"experiments.experiment_id",
			"experiments.name",
			"experiments.lifecycle_stage",
			"experiments.creation_time",
			"COUNT(runs.run_uuid) AS run_count",
			"COALESCE(MAX(experiment_tags.value), '') AS description",
		).
		Where("experiments.namespace_id = ?", ns.ID).
		Where("experiments.lifecycle_stage = ?", database.LifecycleStageActive).
		Joins("LEFT JOIN runs USING(experiment_id)").
		Joins("LEFT JOIN experiment_tags ON experiments.experiment_id = experiment_tags.experiment_id AND"+
			" experiment_tags.key = ?", common.DescriptionTagKey).
		Group("experiments.experiment_id").
		Find(&experiments); tx.Error != nil {
		return fmt.Errorf("error fetching experiments: %w", tx.Error)
	}

	resp := make([]fiber.Map, len(experiments))
	for i, e := range experiments {
		resp[i] = fiber.Map{
			"id":            strconv.Itoa(int(*e.ID)),
			"name":          e.Name,
			"description":   e.Description,
			"archived":      e.LifecycleStage == database.LifecycleStageDeleted,
			"run_count":     e.RunCount,
			"creation_time": float64(e.CreationTime.Int64) / 1000,
		}
	}

	return c.JSON(resp)
}

func GetExperiment(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperiment namespace: %s", ns.Code)

	p := struct {
		ID string `params:"id"`
	}{}

	if err = c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	id, err := strconv.ParseInt(p.ID, 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, fmt.Sprintf("unable to parse experiment id %q: %s", p.ID, err))
	}

	if err := database.DB.Select("ID").First(&database.Experiment{
		ID:          common.GetPointer(int32(id)),
		NamespaceID: ns.ID,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", p.ID, err)
	}

	var exp struct {
		database.Experiment
		RunCount    int
		Description string `gorm:"column:description"`
	}
	if err := database.DB.Model(&database.Experiment{}).
		Select(
			"experiments.experiment_id",
			"experiments.name",
			"experiments.lifecycle_stage",
			"experiments.creation_time",
			"COUNT(runs.run_uuid) AS run_count",
			"COALESCE(MAX(experiment_tags.value), '') AS description",
		).
		Joins("LEFT JOIN runs USING(experiment_id)").
		Joins("LEFT JOIN experiment_tags ON experiments.experiment_id = experiment_tags.experiment_id AND"+
			" experiment_tags.key = ?", common.DescriptionTagKey).
		Where("experiments.namespace_id = ?", ns.ID).
		Where("experiments.experiment_id = ?", id).
		Group("experiments.experiment_id").
		First(&exp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("error fetching experiment %q: %w", p.ID, err)
	}
	return c.JSON(fiber.Map{
		"id":            strconv.Itoa(int(id)),
		"name":          exp.Name,
		"description":   exp.Description,
		"archived":      exp.LifecycleStage == database.LifecycleStageDeleted,
		"run_count":     exp.RunCount,
		"creation_time": float64(exp.CreationTime.Int64) / 1000,
	})
}

func GetExperimentRuns(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperimentRuns namespace: %s", ns.Code)

	q := struct {
		Limit  int    `query:"limit"`
		Offset string `query:"offset"`
	}{}

	if err = c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	p := struct {
		ID string `params:"id"`
	}{}

	if err = c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	id, err := strconv.ParseInt(p.ID, 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, fmt.Sprintf("unable to parse experiment id %q: %s", p.ID, err))
	}

	if err := database.DB.Select("ID").First(&database.Experiment{
		ID:          common.GetPointer(int32(id)),
		NamespaceID: ns.ID,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", p.ID, err)
	}

	tx := database.DB.
		Where("experiment_id = ?", id).
		Order("row_num DESC")

	if q.Limit > 0 {
		tx.Limit(q.Limit)
	}

	if q.Offset != "" {
		run := &database.Run{
			ID: q.Offset,
		}
		if err = database.DB.Select(
			"row_num",
		).First(
			&run,
		).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("unable to find search runs offset %q: %w", q.Offset, err)
		}

		tx.Where("row_num < ?", run.RowNum)
	}

	var sqlRuns []database.Run
	if err := tx.Find(&sqlRuns).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("error fetching runs of experiment %q: %w", p.ID, err)
	}

	runs := make([]fiber.Map, len(sqlRuns))
	for i, r := range sqlRuns {
		runs[i] = fiber.Map{
			"run_id":        r.ID,
			"name":          r.Name,
			"creation_time": float64(r.StartTime.Int64) / 1000,
			"end_time":      float64(r.EndTime.Int64) / 1000,
			"archived":      r.LifecycleStage == database.LifecycleStageDeleted,
		}
	}

	return c.JSON(fiber.Map{
		"id":   p.ID,
		"runs": runs,
	})
}

func GetExperimentActivity(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("GetExperimentActivity namespace: %s", ns.Code)

	tzOffset, err := strconv.Atoi(c.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	p := struct {
		ID string `params:"id"`
	}{}

	if err = c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	id, err := strconv.ParseInt(p.ID, 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, fmt.Sprintf("unable to parse experiment id %q: %s", p.ID, err))
	}

	if err := database.DB.Select(
		"ID",
	).First(&database.Experiment{
		ID:          common.GetPointer(int32(id)),
		NamespaceID: ns.ID,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", p.ID, err)
	}

	var runs []database.Run
	if tx := database.DB.
		Select("runs.start_time", "runs.lifecycle_stage", "runs.status").
		Joins("LEFT JOIN experiments USING(experiment_id)").
		Where("experiments.namespace_id = ?", ns.ID).
		Where("experiments.experiment_id = ?", id).
		Find(&runs); tx.Error != nil {
		return fmt.Errorf("error retrieving runs for experiment %q: %w", p.ID, tx.Error)
	}

	numActiveRuns, numArchivedRuns := 0, 0
	activity := map[string]int{}
	for _, r := range runs {
		key := time.UnixMilli(r.StartTime.Int64).Add(time.Duration(-tzOffset) * time.Minute).Format("2006-01-02T15:00:00")
		activity[key] += 1
		switch {
		case r.LifecycleStage == database.LifecycleStageDeleted:
			numArchivedRuns += 1
		case r.Status == database.StatusRunning:
			numActiveRuns += 1
		}
	}

	return c.JSON(fiber.Map{
		"num_runs":          len(runs),
		"num_archived_runs": numArchivedRuns,
		"num_active_runs":   numActiveRuns,
		"activity_map":      activity,
	})
}

func DeleteExperiment(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteExperiment namespace: %s", ns.Code)

	params := struct {
		ID string `params:"id"`
	}{}

	if err = c.ParamsParser(&params); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	id, err := strconv.ParseInt(params.ID, 10, 32)
	if err != nil {
		return fiber.NewError(
			fiber.StatusUnprocessableEntity, fmt.Sprintf("unable to parse experiment id %q: %s", params.ID, err),
		)
	}

	// validate that requested experiment exists.
	if err := database.DB.Select(
		"ID",
	).Where(
		"experiments.namespace_id = ?", ns.ID,
	).First(&database.Experiment{
		ID: common.GetPointer(int32(id)),
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", params.ID, err)
	}

	// TODO this code should move to service with injected repository
	experimentRepo := repositories.NewExperimentRepository(database.DB)
	if err = experimentRepo.Delete(c.Context(), &models.Experiment{
		ID: common.GetPointer(int32(id)),
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("unable to delete experiment %q: %s", params.ID, err))
	}

	return c.JSON(fiber.Map{
		"id":     params.ID,
		"status": "OK",
	})
}

func UpdateExperiment(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateExperiment namespace: %s", ns.Code)

	params := struct {
		ID string `params:"id"`
	}{}
	if err = c.ParamsParser(&params); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	id, err := strconv.ParseInt(params.ID, 10, 32)
	if err != nil {
		return api.NewBadRequestError("Unable to parse experiment id '%s': %s", params.ID, err)
	}

	var updateRequest request.UpdateExperimentRequest
	if err = c.BodyParser(&updateRequest); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	experimentRepository := repositories.NewExperimentRepository(database.DB)
	tagRepository := repositories.NewTagRepository(database.DB)
	experiment, err := experimentRepository.GetByNamespaceIDAndExperimentID(c.Context(), ns.ID, int32(id))
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError, fmt.Sprintf("unable to find experiment '%s': %s", params.ID, err),
		)
	}
	if experiment == nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("unable to find experiment '%s'", params.ID))
	}
	if updateRequest.Archived != nil {
		if *updateRequest.Archived {
			experiment.LifecycleStage = models.LifecycleStageDeleted
		} else {
			experiment.LifecycleStage = models.LifecycleStageActive
		}
	}

	if updateRequest.Name != nil {
		experiment.Name = *updateRequest.Name
	}

	if updateRequest.Archived != nil || updateRequest.Name != nil {
		if err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := experimentRepository.UpdateWithTransaction(c.Context(), tx, experiment); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError,
				fmt.Sprintf("unable to update experiment %q: %s", params.ID, err))
		}
	}
	if updateRequest.Description != nil {
		description := models.ExperimentTag{
			Key:          common.DescriptionTagKey,
			Value:        *updateRequest.Description,
			ExperimentID: *experiment.ID,
		}
		if err := tagRepository.CreateExperimentTag(c.Context(), &description); err != nil {
			return err
		}
	}

	return c.JSON(fiber.Map{
		"id":     params.ID,
		"status": "OK",
	})
}
