package fixtures

import (
	"context"
	"time"

	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/api/aim/response"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// ProjectFixtures represents data fixtures object.
type ProjectFixtures struct {
	baseFixtures
}

// NewProjectFixtures creates new instance of ProjectFixtures.
func NewProjectFixtures(databaseDSN string) (*ProjectFixtures, error) {
	db, err := database.MakeDBProvider(
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
		baseFixtures: baseFixtures{db: db.GormDB()},
	}, nil
}

// GetProject returns a GetProjectResponse.
func (f *ProjectFixtures) GetProject(ctx context.Context) *response.GetProjectResponse {
	return &response.GetProjectResponse{
		Name:             "FastTrackML",
		Path:             f.db.Dialector.Name(),
		Description:      "",
		TelemetryEnabled: float64(0),
	}
}

// GetProjectActivity returns a summary of ProjectActivityResponse summary.
func (f *ProjectFixtures) GetProjectActivity(
	ctx context.Context,
) (*response.ProjectActivityResponse, error) {
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

	activity := map[string]int{time.Now().Format("2006-01-02T15:00:00"): int(numRuns)}

	return &response.ProjectActivityResponse{
		NumExperiments:  float64(numExperiments),
		NumRuns:         float64(numRuns),
		NumArchivedRuns: float64(numArchivedRuns),
		NumActiveRuns:   float64(numActiveRuns),
		ActivityMap:     activity,
	}, nil
}
