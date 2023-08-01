package fixtures

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rotisserie/eris"
 
	"github.com/G-Research/fasttrackml/pkg/database"
)

// ProjectFixtures represents data fixtures object.
type ProjectFixtures struct {
	baseFixtures
}

// NewProjectFixtures creates new instance of ProjectFixtures.
func NewProjectFixtures(databaseDSN string) (*ProjectFixtures, error) {
	db, err := database.ConnectDB(
		databaseDSN,
		1*time.Second,
		20,
		false,
		false,
		"",
	)
	if err != nil {
		return nil, eris.Wrap(err, "error connection to database")
	}
	return &ProjectFixtures{
		baseFixtures: baseFixtures{db: db.DB},
	}, nil
}

func (f *ProjectFixtures) GetProject(ctx context.Context) *fiber.Map {
	return &fiber.Map{
		"name":              "FastTrackML",
		"path":              database.DB.DSN(),
		"description":       "",
		"telemetry_enabled": float64(0),
	}
}

func (f *ProjectFixtures) GetProjectActivity(
	ctx context.Context,
) (*fiber.Map, error) {
	var numExperiments int64
	if err := f.db.WithContext(ctx).
		Table("experiments").
		Count(&numExperiments).Error; err != nil {
		return nil, eris.Wrapf(err, "error counting experiments")
	}

	var numRuns int64
	if err := f.db.WithContext(ctx).
		Table("runs").
		Count(&numRuns).Error; err != nil {
		return nil, eris.Wrapf(err, "error counting runs")
	}

	var numActiveRuns int64
	if err := f.db.WithContext(ctx).
		Table("runs").
		Where("status = ?", database.StatusRunning).
		Count(&numActiveRuns).Error; err != nil {
		return nil, eris.Wrapf(err, "error counting active runs")
	}

	var numArchivedRuns int64
	if err := f.db.WithContext(ctx).
		Table("runs").
		Where("lifecycle_stage = ?", database.LifecycleStageDeleted).
		Count(&numArchivedRuns).Error; err != nil {
		return nil, eris.Wrapf(err, "error counting archived runs")
	}

    activity := map[string]int{}
    key := time.Now().Format("2006-01-02T15:00:00")
	activity[key] = int(numRuns)

	return &fiber.Map{
		"num_experiments":   float64(numExperiments),
		"num_runs":          float64(numRuns),
		"num_archived_runs": float64(numArchivedRuns),
		"num_active_runs":   float64(numActiveRuns),
		"activity_map":      activity,
	}, nil
}
