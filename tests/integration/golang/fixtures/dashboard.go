package fixtures

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"

	"github.com/G-Research/fasttrackml/pkg/database"
)

// DashboardFixtures represents data fixtures object.
type DashboardFixtures struct {
	baseFixtures
	*database.DbInstance
}

// NewDashboardFixtures creates new instance of DashboardFixtures.
func NewDashboardFixtures(databaseDSN string) (*DashboardFixtures, error) {
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
	return &DashboardFixtures{
		baseFixtures: baseFixtures{db: db.DB},
		DbInstance:   db,
	}, nil
}

// CreateDashboard creates a new test Dashboard.
func (f DashboardFixtures) CreateDashboard(
	ctx context.Context, dashboard *database.Dashboard,
) (*database.Dashboard, error) {
	if err := f.db.WithContext(ctx).Create(dashboard).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test dashboard")
	}
	return dashboard, nil
}

// CreateDashboards creates some num dashboards belonging to the experiment
func (f DashboardFixtures) CreateDashboards(
	ctx context.Context, num int, appId *uuid.UUID,
) ([]*database.Dashboard, error) {
	var dashboards []*database.Dashboard
	// create dashboards for the experiment
	for i := 0; i < num; i++ {
		dashboard := &database.Dashboard{
			Base: database.Base{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
			},
			Name:  "dashboard-exp",
			Description: "dashboard for experiment",
			AppID: appId,
		}
		dashboard, err := f.CreateDashboard(ctx, dashboard)
		if err != nil {
			return nil, err
		}
		dashboards = append(dashboards, dashboard)
	}
	return dashboards, nil
}