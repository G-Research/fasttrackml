package fixtures

import (
	"context"

	"github.com/rotisserie/eris"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// AppFixtures represents data fixtures object.
type AppFixtures struct {
	baseFixtures
}

// NewAppFixtures creates new instance of AppFixtures.
func NewAppFixtures(db *gorm.DB) (*AppFixtures, error) {
	return &AppFixtures{
		baseFixtures: baseFixtures{db: db},
	}, nil
}

// CreateApp creates a new test App.
func (f AppFixtures) CreateApp(
	ctx context.Context, app *database.App,
) (*database.App, error) {
	if err := f.db.WithContext(ctx).Create(app).Error; err != nil {
		return nil, eris.Wrap(err, "error creating test app")
	}
	return app, nil
}

// CreateApps creates some num apps belonging to the experiment.
func (f AppFixtures) CreateApps(ctx context.Context, namespace *models.Namespace, num int) ([]*database.App, error) {
	var apps []*database.App
	// create apps for the experiment
	for i := 0; i < num; i++ {
		app := &database.App{
			Type:        "mpi",
			State:       database.AppState{},
			NamespaceID: namespace.ID,
		}
		app, err := f.CreateApp(ctx, app)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// GetApps fetches all apps which are not archived.
func (f AppFixtures) GetApps(
	ctx context.Context,
) ([]database.App, error) {
	apps := []database.App{}
	if err := f.db.WithContext(ctx).
		Where("NOT is_archived").
		Find(&apps).Error; err != nil {
		return nil, eris.Wrapf(err, "error getting 'app' entities")
	}
	return apps, nil
}
