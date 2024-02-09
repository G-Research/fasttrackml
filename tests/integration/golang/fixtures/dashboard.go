package fixtures

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// DashboardFixtures represents data fixtures object.
type DashboardFixtures struct {
	baseFixtures
}

// NewDashboardFixtures creates new instance of DashboardFixtures.
func NewDashboardFixtures(db *gorm.DB) (*DashboardFixtures, error) {
	return &DashboardFixtures{
		baseFixtures: baseFixtures{db: db},
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
	ctx context.Context, namespace *models.Namespace, num int,
) ([]*database.Dashboard, error) {
	var dashboards []*database.Dashboard
	// create dashboards for the experiment
	for i := 0; i < num; i++ {
		dashboard, err := f.CreateDashboard(ctx, &database.Dashboard{
			Name:        fmt.Sprintf("dashboard-%d", i+1),
			Description: fmt.Sprintf("Dashboard %d", i+1),
			App: database.App{
				Type:        "mpi",
				State:       database.AppState{},
				NamespaceID: namespace.ID,
			},
		})
		if err != nil {
			return nil, err
		}
		dashboards = append(dashboards, dashboard)
	}
	return dashboards, nil
}

// GetDashboardByID returns database.Dashboard entity by its ID.
func (f DashboardFixtures) GetDashboardByID(ctx context.Context, dashboardID string) (*database.Dashboard, error) {
	var dashboard database.Dashboard
	if err := f.db.WithContext(ctx).Where(
		"id = ?", dashboardID,
	).Where(
		"NOT is_archived",
	).Find(
		&dashboard,
	).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting 'dashboard' entity by id: %s", dashboardID)
	}
	return &dashboard, nil
}

// GetDashboards fetches all dashboards which are not archived
func (f DashboardFixtures) GetDashboards(
	ctx context.Context,
) ([]database.Dashboard, error) {
	var dashboards []database.Dashboard
	if err := f.db.WithContext(ctx).Where(
		"NOT is_archived",
	).Find(
		&dashboards,
	).Error; err != nil {
		return nil, eris.Wrap(err, "error getting 'dashboard' entities")
	}
	return dashboards, nil
}
