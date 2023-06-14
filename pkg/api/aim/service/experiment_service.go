package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/G-Research/fasttrackml/pkg/database"
	"github.com/G-Research/fasttrackml/pkg/models"
	"github.com/G-Research/fasttrackml/pkg/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ExperimentService struct {
	experimentRepository repositories.ExperimentRepositoryProvider
}

func NewExperimentService(experimentRepo repositories.ExperimentRepositoryProvider) *ExperimentService {
	return &ExperimentService{
		experimentRepository: experimentRepo,
	}
}

func (svc ExperimentService) GetExperiments(ctx context.Context) (*[]models.Experiment, error) {
	return svc.experimentRepository.List(ctx)
}

func (svc ExperimentService) GetExperiment(ctx context.Context, id int32) (*models.Experiment, error) {

	return svc.experimentRepository.GetByID(ctx, id)
}

func (svc ExperimentService) GetExperimentRuns(ctx context.Context, id int32, limit int, offset string) error {

	if tx := database.DB.Select("ID").First(&database.Experiment{
		ID: &id,
	}); tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", id, tx.Error)
	}

	tx := database.DB.
		Where("experiment_id = ?", id).
		Order("row_num DESC")

	if limit > 0 {
		tx.Limit(limit)
	}

	if offset != "" {
		run := &database.Run{
			ID: offset,
		}
		if tx := database.DB.Select("row_num").First(&run); tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("unable to find search runs offset %q: %w", offset, tx.Error)
		}

		tx.Where("row_num < ?", run.RowNum)
	}

	var sqlRuns []database.Run
	tx.Find(&sqlRuns)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("error fetching runs of experiment %q: %w", id, tx.Error)
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
	return nil
}

func (svc ExperimentService) GetExperimentActivity(c *fiber.Ctx) error {
	tzOffset, err := strconv.Atoi(c.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	p := struct {
		ID string `params:"id"`
	}{}

	if err := c.ParamsParser(&p); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	id, err := strconv.ParseInt(p.ID, 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, fmt.Sprintf("unable to parse experiment id %q: %s", p.ID, err))
	}
	id32 := int32(id)

	if tx := database.DB.Select("ID").First(&database.Experiment{
		ID: &id32,
	}); tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return fiber.ErrNotFound
		}
		return fmt.Errorf("unable to find experiment %q: %w", p.ID, tx.Error)
	}

	var runs []database.Run
	if tx := database.DB.
		Select("StartTime", "LifecycleStage", "Status").
		Where("experiment_id = ?", id).
		Find(&runs); tx.Error != nil {
		return fmt.Errorf("error retrieving runs for experiment %q: %w", p.ID, tx.Error)
	}

	numArchivedRuns := 0
	numActiveRuns := 0
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
