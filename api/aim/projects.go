package aim

import (
	"fmt"
	"strconv"
	"time"

	"github.com/G-Resarch/fasttrack/database"

	"github.com/gofiber/fiber/v2"
)

func GetProject(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name":              "fasttrack",
		"path":              database.DB.DSN(),
		"description":       "",
		"telemetry_enabled": 0,
	})
}

func GetProjectActivity(c *fiber.Ctx) error {
	tzOffset, err := strconv.Atoi(c.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	var numExperiments int64
	if tx := database.DB.Model(&database.Experiment{}).Where("lifecycle_stage = ?", database.LifecycleStageActive).Count(&numExperiments); tx.Error != nil {
		return fmt.Errorf("error counting experiments: %w", tx.Error)
	}

	var runs []database.Run
	if tx := database.DB.Select("StartTime", "LifecycleStage", "Status").Find(&runs); tx.Error != nil {
		return fmt.Errorf("error retrieving runs: %w", tx.Error)
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
		"num_experiments":   numExperiments,
		"num_runs":          len(runs),
		"num_archived_runs": numArchivedRuns,
		"num_active_runs":   numActiveRuns,
		"activity_map":      activity,
	})
}

// TODO
func GetProjectPinnedSequences(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"sequences": []string{},
	})
}

// TODO
func UpdateProjectPinnedSequences(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"sequences": []string{},
	})
}

func GetProjectParams(c *fiber.Ctx) error {
	q := struct {
		ExcludeParams bool     `query:"exclude_params"`
		Sequences     []string `query:"sequence"`
	}{}

	if err := c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	resp := fiber.Map{}

	if !q.ExcludeParams {
		var paramKeys []string
		if tx := database.DB.Model(&database.Param{}).Distinct().Pluck("Key", &paramKeys); tx.Error != nil {
			return fmt.Errorf("error retrieving param keys: %w", tx.Error)
		}

		params := make(map[string]map[string]string, len(paramKeys))
		for _, p := range paramKeys {
			params[p] = map[string]string{
				"__example_type__": "<class 'str'>",
			}
		}

		resp["params"] = params
	}

	if len(q.Sequences) == 0 {
		q.Sequences = []string{
			"metric",
			"images",
			"texts",
			"figures",
			"distributions",
			"audios",
		}
	}

	for _, s := range q.Sequences {
		switch s {
		case "images", "texts", "figures", "distributions", "audios":
			resp[s] = fiber.Map{}
		case "metric":
			var metricKeys []string
			if tx := database.DB.Model(&database.Metric{}).Distinct().Pluck("Key", &metricKeys); tx.Error != nil {
				return fmt.Errorf("error retrieving metric keys: %w", tx.Error)
			}

			metrics := make(map[string][]fiber.Map, len(metricKeys))
			for _, m := range metricKeys {
				metrics[m] = []fiber.Map{{}}
			}

			resp[s] = metrics
		default:
			return fmt.Errorf("%q is not a valid Sequence", s)
		}
	}

	return c.JSON(resp)
}

func GetProjectStatus(c *fiber.Ctx) error {
	return c.JSON("up-to-date")
}
