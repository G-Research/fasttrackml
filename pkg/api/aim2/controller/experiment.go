package controller

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim2/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/convertors"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/aim2/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/common"
	"github.com/G-Research/fasttrackml/pkg/common/api"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
)

func (c Controller) GetExperiments(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
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

	return ctx.JSON(resp)
}

func (c Controller) GetExperiment(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperiment namespace: %s", ns.Code)

	req := request.GetExperimentRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := database.DB.Select("ID").First(&database.Experiment{
		ID:          common.GetPointer(req.ID),
		NamespaceID: ns.ID,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", req.ID, err)
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
		Where("experiments.experiment_id = ?", req.ID).
		Group("experiments.experiment_id").
		First(&exp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("error fetching experiment %q: %w", req.ID, err)
	}
	return ctx.JSON(fiber.Map{
		"id":            fmt.Sprintf("%d", req.ID),
		"name":          exp.Name,
		"description":   exp.Description,
		"archived":      exp.LifecycleStage == database.LifecycleStageDeleted,
		"run_count":     exp.RunCount,
		"creation_time": float64(exp.CreationTime.Int64) / 1000,
	})
}

func (c Controller) GetExperimentRuns(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("getExperimentRuns namespace: %s", ns.Code)

	req := request.GetExperimentRunsRequest{}
	if err = ctx.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := database.DB.Select("ID").First(&database.Experiment{
		ID:          common.GetPointer(req.ID),
		NamespaceID: ns.ID,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", req.ID, err)
	}

	tx := database.DB.
		Where("experiment_id = ?", req.ID).
		Order("row_num DESC")

	if req.Limit > 0 {
		tx.Limit(req.Limit)
	}

	if req.Offset != "" {
		run := &database.Run{
			ID: req.Offset,
		}
		if err = database.DB.Select(
			"row_num",
		).First(
			&run,
		).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("unable to find search runs offset %q: %w", req.Offset, err)
		}

		tx.Where("row_num < ?", run.RowNum)
	}

	var sqlRuns []database.Run
	if err := tx.Find(&sqlRuns).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("error fetching runs of experiment %q: %w", req.ID, err)
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

	return ctx.JSON(fiber.Map{
		"id":   req.ID,
		"runs": runs,
	})
}

func (c Controller) GetExperimentActivity(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("GetExperimentActivity namespace: %s", ns.Code)

	tzOffset, err := strconv.Atoi(ctx.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	req := request.GetExperimentActivityRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := database.DB.Select(
		"ID",
	).First(&database.Experiment{
		ID:          common.GetPointer(req.ID),
		NamespaceID: ns.ID,
	}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", req.ID, err)
	}

	var runs []database.Run
	if tx := database.DB.
		Select("runs.start_time", "runs.lifecycle_stage", "runs.status").
		Joins("LEFT JOIN experiments USING(experiment_id)").
		Where("experiments.namespace_id = ?", ns.ID).
		Where("experiments.experiment_id = ?", req.ID).
		Find(&runs); tx.Error != nil {
		return fmt.Errorf("error retrieving runs for experiment %q: %w", req.ID, tx.Error)
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

	return ctx.JSON(fiber.Map{
		"num_runs":          len(runs),
		"num_archived_runs": numArchivedRuns,
		"num_active_runs":   numActiveRuns,
		"activity_map":      activity,
	})
}

func (c Controller) DeleteExperiment(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("deleteExperiment namespace: %s", ns.Code)

	req := request.DeleteExperimentRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	// validate that requested experiment exists and is not a default experiment.
	experiment := database.Experiment{}
	if err := database.DB.Select(
		"ID", "Name",
	).Where(
		"experiments.experiment_id = ?", req.ID,
	).Where(
		"experiments.namespace_id = ?", ns.ID,
	).First(&experiment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", req.ID, err)
	}

	if experiment.IsDefault(ns) {
		return fiber.NewError(fiber.StatusBadRequest, "unable to delete default experiment")
	}

	// TODO this code should move to service with injected repository
	experimentRepo := repositories.NewExperimentRepository(database.DB)
	if err = experimentRepo.Delete(ctx.Context(), &models.Experiment{
		ID: common.GetPointer(req.ID),
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("unable to delete experiment %q: %s", req.ID, err))
	}

	return ctx.JSON(fiber.Map{
		"id":     req.ID,
		"status": "OK",
	})
}

func (c Controller) UpdateExperiment(ctx *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(ctx.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}
	log.Debugf("updateExperiment namespace: %s", ns.Code)

	req := request.UpdateExperimentRequest{}
	if err = ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err = ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	experimentRepository := repositories.NewExperimentRepository(database.DB)
	experiment, err := experimentRepository.GetByNamespaceIDAndExperimentID(ctx.Context(), ns.ID, req.ID)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError, fmt.Sprintf("unable to find experiment '%d': %s", req.ID, err),
		)
	}
	if experiment == nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("unable to find experiment '%d'", req.ID))
	}

	experiment = convertors.ConvertUpdateExperimentToDBModel(&req, experiment)
	if req.Archived != nil || req.Name != nil {
		if err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := experimentRepository.UpdateWithTransaction(ctx.Context(), tx, experiment); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError,
				fmt.Sprintf("unable to update experiment %q: %s", req.ID, err))
		}
	}
	if req.Description != nil {
		tagRepository := repositories.NewTagRepository(database.DB)
		if err := tagRepository.CreateExperimentTag(ctx.Context(), &models.ExperimentTag{
			Key:          common.DescriptionTagKey,
			Value:        *req.Description,
			ExperimentID: *experiment.ID,
		}); err != nil {
			return err
		}
	}

	return ctx.JSON(fiber.Map{
		"id":     req.ID,
		"status": "OK",
	})
}
